package db

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/x/posts"
	"github.com/desmos-labs/djuno/types"
	"github.com/rs/zerolog/log"
)

// PostRow represents a single PostgreSQL row containing the data of a Post
type RegisteredReactionRow struct {
	ReactionID uint64  `db:"id"`
	OwnerID    *uint64 `db:"owner_id"`
	ShortCode  string  `db:"short_code"`
	Value      string  `db:"value"`
	Subspace   string  `db:"subspace"`
}

// ConvertPostRow takes the given postRow and userRow and merges the data contained inside them to create a Post.
func ConvertReactionRow(reactionRow RegisteredReactionRow, userRow *UserRow) (*posts.Reaction, error) {

	// Parse the creator
	creator, err := sdk.AccAddressFromBech32(userRow.Address)
	if err != nil {
		return nil, err
	}

	// Create the reaction
	reaction := posts.NewReaction(creator, reactionRow.ShortCode, reactionRow.Value, reactionRow.Subspace)
	return &reaction, nil
}

// SaveReaction allows to save the given reaction into the database.
func (db DesmosDb) SaveReaction(reaction *types.PostReaction) error {
	owner, err := db.SaveUserIfNotExisting(reaction.User)
	if err != nil {
		return err
	}

	log.Info().
		Str("post_id", reaction.PostID.String()).
		Str("value", reaction.Value).
		Str("short_code", reaction.ShortCode).
		Str("user", reaction.User.String()).
		Msg("saving reaction")

	statement := `INSERT INTO reaction (post_id, owner_id, short_code, value) VALUES ($1, $2, $3, $4)`
	_, err = db.Sql.Exec(statement, reaction.PostID, owner.Id, reaction.ShortCode, reaction.Value)
	return err
}

// RemoveReaction allows to remove an already existing reaction from the database.
func (db DesmosDb) RemoveReaction(reaction *types.PostReaction) error {
	owner, err := db.SaveUserIfNotExisting(reaction.User)
	if err != nil {
		return err
	}

	log.Info().
		Str("post_id", reaction.PostID.String()).
		Str("value", reaction.Value).
		Str("short_code", reaction.ShortCode).
		Str("user", reaction.User.String()).
		Msg("removing reaction")

	statement := `DELETE FROM reaction WHERE post_id = $1 AND owner_id = $2 AND short_code = $3`
	_, err = db.Sql.Exec(statement, reaction.PostID.String(), owner.Id, reaction.ShortCode)
	return err
}

// GetRegisteredReactionByCodeOrValue allows to get a registered reaction by its shortcode or
// value and the subspace for which it has been registered.
func (db DesmosDb) GetRegisteredReactionByCodeOrValue(codeOrValue string, subspace string) (*posts.Reaction, error) {
	postSqlStatement := `SELECT * FROM registered_reactions WHERE (short_code = $1 OR value = $1) AND subspace = $2`

	var rows []RegisteredReactionRow
	err := db.sqlx.Select(&rows, postSqlStatement, codeOrValue, subspace)
	if err != nil {
		return nil, err
	}

	// No post found
	if len(rows) == 0 {
		return nil, nil
	}

	reactionRow := rows[0]

	// Find the user
	userRow, err := db.GetUserById(reactionRow.OwnerID)
	if err != nil {
		return nil, err
	}

	return ConvertReactionRow(reactionRow, userRow)
}

// RegisterReaction allows to register into the database the given reaction.
func (db DesmosDb) RegisterReactionIfNotPresent(reaction posts.Reaction) (*posts.Reaction, error) {
	react, err := db.GetRegisteredReactionByCodeOrValue(reaction.ShortCode, reaction.Subspace)
	if err != nil {
		return nil, err
	}

	// If the reaction exists do nothing
	if react != nil {
		return react, nil
	}

	// Save the owner
	owner, err := db.SaveUserIfNotExisting(reaction.Creator)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("value", reaction.Value).
		Str("short_code", reaction.ShortCode).
		Str("creator", reaction.Creator.String()).
		Msg("registering reaction")

	// Save the reaction
	statement := `INSERT INTO registered_reactions (owner_id, short_code, value, subspace) VALUES ($1, $2, $3, $4)`
	_, err = db.Sql.Exec(statement, owner.Id, reaction.ShortCode, reaction.Value, reaction.Subspace)
	return &reaction, err
}
