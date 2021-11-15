package profiles

import (
	"context"
	"fmt"

	"github.com/forbole/juno/v2/node/remote"

	"github.com/desmos-labs/djuno/types"

	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	oracletypes "github.com/desmos-labs/desmos/v2/x/oracle/types"
	profilestypes "github.com/desmos-labs/desmos/v2/x/profiles/types"
)

// packetHandler defines a function that handles a packet.
// It returns true iff it was able to handle the packet, and an error if something goes wrong.
type packetHandler = func(height int64, packet channeltypes.Packet) (bool, error)

// handlePacket tries handling the given packet that was received at the given height
func (m *Module) handlePacket(height int64, packet channeltypes.Packet) error {
	// Try handling the packet
	handlers := []packetHandler{
		m.handleLinkChainAccountPacketData,
		m.handleOracleRequestPacketData,
		m.handleOracleResponsePacketData,
	}

	for _, handler := range handlers {
		handled, err := handler(height, packet)
		if handled {
			return err
		}
	}

	return fmt.Errorf("cannot handle packet directed to port %s and channel %s", packet.DestinationPort, packet.DestinationChannel)
}

// handleLinkChainAccountPacketData tries handling the given packet as it contains a LinkChainAccountPacketData
// instance. This is done to store chain links that are created using IBC.
func (m *Module) handleLinkChainAccountPacketData(height int64, packet channeltypes.Packet) (bool, error) {
	// Try reading the packet data
	var packetData profilestypes.LinkChainAccountPacketData
	err := m.cdc.UnmarshalJSON(packet.GetData(), &packetData)
	if err != nil {
		return false, nil
	}

	var sourceAddr profilestypes.AddressData
	err = m.cdc.UnpackAny(packetData.SourceAddress, &sourceAddr)
	if err != nil {
		return true, fmt.Errorf("error while deserializing source address: %s", err)
	}

	// Get the link from the chain
	res, err := m.profilesClient.UserChainLink(
		context.Background(),
		&profilestypes.QueryUserChainLinkRequest{
			User:      packetData.DestinationAddress,
			ChainName: packetData.SourceChainConfig.Name,
			Target:    sourceAddr.GetValue(),
		},
		remote.GetHeightRequestHeader(height),
	)
	if err != nil {
		return true, err
	}

	// Save the chain link
	return true, m.db.SaveChainLink(types.NewChainLink(res.Link, height))
}

// handleOracleRequestPacketData tries handling the given packet as it contains a OracleRequestPacketData
// instance. This is done in order to update existing application links when their state changes after
// Band Protocol ends the verification process.
func (m *Module) handleOracleRequestPacketData(height int64, packet channeltypes.Packet) (bool, error) {
	var data oracletypes.OracleRequestPacketData
	if err := m.cdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return false, nil
	}

	res, err := m.profilesClient.ApplicationLinkByClientID(
		context.Background(),
		profilestypes.NewQueryApplicationLinkByClientIDRequest(data.ClientID),
		remote.GetHeightRequestHeader(height),
	)
	if err != nil {
		return true, fmt.Errorf("error while getting application link by client id: %s", err)
	}

	return true, m.db.SaveApplicationLink(types.NewApplicationLink(res.Link, height))
}

// handleOracleResponsePacketData tries handling the given packet as it contains a OracleResponsePacketData
// instance. This is done in order to update existing application links when their state changes after
// Band Protocol ends the verification process.
func (m *Module) handleOracleResponsePacketData(height int64, packet channeltypes.Packet) (bool, error) {
	var data oracletypes.OracleResponsePacketData
	if err := m.cdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return false, nil
	}

	res, err := m.profilesClient.ApplicationLinkByClientID(
		context.Background(),
		profilestypes.NewQueryApplicationLinkByClientIDRequest(data.ClientID),
		remote.GetHeightRequestHeader(height),
	)
	if err != nil {
		return true, fmt.Errorf("error while getting application link by client id: %s", err)
	}

	return true, m.db.SaveApplicationLink(types.NewApplicationLink(res.Link, height))
}
