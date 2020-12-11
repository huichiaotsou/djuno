package notifications_test

import (
	"testing"

	poststypes "github.com/desmos-labs/desmos/x/posts/types"
	"github.com/desmos-labs/djuno/notifications"
	"github.com/stretchr/testify/require"
)

func TestGetPostMentions(t *testing.T) {
	result, err := notifications.GetPostMentions(poststypes.Post{
		Message: `Hello @desmos1p7c8h59nrc8e5hxvgvu2g7tpp0xwn4mzevzgg7! 
				  How is it going @desmos1p7ad878nealg249qkkdl9ldxrllst23lklngcx?`,
	})
	require.NoError(t, err)

	expected := []string{
		"desmos1p7c8h59nrc8e5hxvgvu2g7tpp0xwn4mzevzgg7",
		"desmos1p7ad878nealg249qkkdl9ldxrllst23lklngcx",
	}
	require.Len(t, result, len(expected))
	for index, address := range result {
		require.Equal(t, expected[index], address)
	}
}
