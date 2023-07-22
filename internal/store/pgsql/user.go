package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"be-wedding/internal/store"

	"github.com/google/uuid"
)

type User struct {
	db *sql.DB
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

const userInsertQuery = `INSERT INTO
users(
	id, invitation_id, wa_number, status, qr_image, created_at
) values(
	$1, $2, $3, $4, $5, $6
)
`

const invitationStatusUpdateQuery = `UPDATE invitations
	SET status = $2, updated_at = $3
	WHERE id = $1
`

func (s *User) Insert(ctx context.Context, user *store.UserData) error {
	insertStmt, err := s.db.PrepareContext(ctx, userInsertQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.NewString()
	createdAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
		userID, user.InvitationID, user.WhatsAppNumber, store.UserStatusNewlyCreated,
		user.QRImageName, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	// Update Invitation Status
	if user.InvitationType == store.InvitationTypeSingle {

		invitationUpdatedAt := time.Now().UTC()
		updateInvitationStatusStmt, err := s.db.PrepareContext(ctx, invitationStatusUpdateQuery)
		_, err = tx.StmtContext(ctx, updateInvitationStatusStmt).ExecContext(ctx,
			user.InvitationID, store.InvitationStatusUsed, invitationUpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update invitation status: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	user.ID = userID
	user.CreatedAt = createdAt

	return nil

}

func (s *User) InsertWithID(ctx context.Context, user *store.UserData) error {
	insertStmt, err := s.db.PrepareContext(ctx, userInsertQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	createdAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
		user.ID, user.InvitationID, user.WhatsAppNumber, store.UserStatusNewlyCreated,
		user.QRImageName, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	// Update Invitation Status
	if user.InvitationType == store.InvitationTypeSingle {

		invitationUpdatedAt := time.Now().UTC()
		updateInvitationStatusStmt, err := s.db.PrepareContext(ctx, invitationStatusUpdateQuery)
		_, err = tx.StmtContext(ctx, updateInvitationStatusStmt).ExecContext(ctx,
			user.InvitationID, store.InvitationStatusUsed, invitationUpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to update invitation status: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	user.CreatedAt = createdAt

	return nil

}

const userUpdateQuery = `UPDATE users
	SET name = $2, status = $3, updated_at = $4
	WHERE id = $1
`

func (s *User) Update(ctx context.Context, user *store.UserData) error {
	updateStmt, err := s.db.PrepareContext(ctx, userUpdateQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		user.ID, user.Name, store.UserStatusInfoCompleted, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	user.Status = store.UserStatusInfoCompleted
	user.UpdatedAt = sql.NullTime{Time: updatedAt, Valid: true}

	return nil

}

const insertCommentQuery = `INSERT INTO
user_comments(
	id, user_id, comment, created_at
) values(
	$1, $2, $3, $4
)
`

func (s *User) InsertComment(ctx context.Context, userComment *store.UserCommentData) error {
	insertStmt, err := s.db.PrepareContext(ctx, insertCommentQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.NewString()
	createdAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
		userID, userComment.UserID, userComment.Comment, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	userComment.ID = userID
	userComment.CreatedAt = createdAt

	return nil

}

const userFindAllCommentQuery = `SELECT uc.id, uc.user_id, u.name, uc.comment, uc.created_at, count(ucl.id) 
	FROM user_comments uc
	LEFT JOIN users u
	ON uc.user_id = u.id
	LEFT JOIN user_comment_likes ucl 
	ON uc.id = ucl.comment_id 
	`

func (s *User) FindAllComment(ctx context.Context, startDate string, endDate string) ([]*store.UserCommentData, error) {
	userCommentList := []*store.UserCommentData{}
	var queryKeys []string
	var queryParams []interface{}

	query := userFindAllCommentQuery

	if startDate != "" {
		queryKeys = append(queryKeys, "StartDate")
		queryParams = append(queryParams, startDate)
	}

	if endDate != "" {
		queryKeys = append(queryKeys, "EndDate")
		queryParams = append(queryParams, endDate)
	}

	for index, key := range queryKeys {
		if index == 0 {
			query = query + "WHERE "
		} else {
			query = query + "AND "
		}

		switch key {
		case "StartDate":
			query = query + fmt.Sprintf(`uc.created_at >= $%d `, index+1)
		case "EndDate":
			query = query + fmt.Sprintf(`uc.created_at <= $%d `, index+1)
		}
	}

	query = query + `GROUP BY uc.id, uc.user_id, u."name", uc."comment", uc.created_at
ORDER BY count(ucl.id) DESC, uc.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userComment := &store.UserCommentData{}
		err := rows.Scan(
			&userComment.ID,
			&userComment.UserID,
			&userComment.UserName,
			&userComment.Comment,
			&userComment.CreatedAt,
			&userComment.Like,
		)
		if err != nil {
			return nil, err
		}
		userCommentList = append(userCommentList, userComment)
	}

	return userCommentList, nil
}

func (r *User) FindAllCommentPagination(ctx context.Context, offset int, limit int, startDate string, endDate string) ([]*store.UserCommentData, error) {
	userCommentList := []*store.UserCommentData{}
	var queryKeys []string
	var queryParams []interface{}

	queryParams = append(queryParams, offset, limit)

	query := userFindAllCommentQuery

	if startDate != "" {
		queryKeys = append(queryKeys, "StartDate")
		queryParams = append(queryParams, startDate)
	}

	if endDate != "" {
		queryKeys = append(queryKeys, "EndDate")
		queryParams = append(queryParams, endDate)
	}

	for index, key := range queryKeys {
		if index == 0 {
			query = query + "WHERE "
		} else {
			query = query + "AND "
		}

		// +3 is used because $1 and $2 are already used for pagination
		switch key {
		case "StartDate":
			query = query + fmt.Sprintf(`uc.created_at >= $%d `, index+3)
		case "EndDate":
			query = query + fmt.Sprintf(`uc.created_at <= $%d `, index+3)
		}
	}

	query = query + `GROUP BY uc.id, uc.user_id, u."name", uc."comment", uc.created_at
ORDER BY count(ucl.id) DESC, uc.created_at DESC LIMIT $2 OFFSET $1 `

	rows, err := r.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userComment := &store.UserCommentData{}
		err := rows.Scan(
			&userComment.ID,
			&userComment.UserID,
			&userComment.UserName,
			&userComment.Comment,
			&userComment.CreatedAt,
			&userComment.Like,
		)
		if err != nil {
			log.Println("user sql repo scan error: %w", err)
			return nil, err
		}
		userCommentList = append(userCommentList, userComment)
	}

	return userCommentList, nil
}

const userCommentLikeFindOneByUserAndCommentIDQuery = `SELECT id
	FROM user_comment_likes
	WHERE user_id = $1 AND comment_id = $2
`

const userInsertCommentLikeQuery = `INSERT INTO
user_comment_likes (
	id, user_id, comment_id, created_at
) values(
	$1, $2, $3, $4
)
`

const userDeleteCommentLikeQuery = `DELETE FROM user_comment_likes
	WHERE user_id = $1 AND comment_id = $2
`

func (s *User) LikeUnlikeComment(ctx context.Context, userID string, commentID string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	userCommentLike := &store.UserCommentLikeData{}
	var isLiked bool

	row := s.db.QueryRowContext(ctx, userCommentLikeFindOneByUserAndCommentIDQuery, userID, commentID)

	err = row.Scan(
		&userCommentLike.ID,
	)
	if err != nil { // like
		insertStmt, err := s.db.PrepareContext(ctx, userInsertCommentLikeQuery)
		if err != nil {
			return false, err
		}

		userLikeCommentID := uuid.NewString()
		createdAt := time.Now().UTC()
		_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
			userLikeCommentID, userID, commentID, createdAt,
		)
		if err != nil {
			return false, fmt.Errorf("failed to insert: %w", err)
		}
		isLiked = true
	} else { // unlike
		deleteStmt, err := s.db.PrepareContext(ctx, userDeleteCommentLikeQuery)
		if err != nil {
			return false, err
		}

		_, err = tx.StmtContext(ctx, deleteStmt).ExecContext(ctx,
			userID, commentID,
		)
		if err != nil {
			return false, fmt.Errorf("failed to delete: %w", err)
		}
		isLiked = false
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit: %w", err)
	}

	return isLiked, nil

}

const userFindLikedCommentByUserIDQuery = `SELECT ucl.id, ucl.user_id, ucl.comment_id
	FROM user_comment_likes ucl
	WHERE ucl.user_id = $1
	`

func (s *User) FindLikedCommentOnlyByUserID(ctx context.Context, userID string) ([]*store.UserCommentLikeData, error) {
	userCommentLikeList := []*store.UserCommentLikeData{}

	rows, err := s.db.QueryContext(ctx, userFindLikedCommentByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userCommentLike := &store.UserCommentLikeData{}
		err := rows.Scan(
			&userCommentLike.ID,
			&userCommentLike.UserID,
			&userCommentLike.CommentID,
		)
		if err != nil {
			return nil, err
		}
		userCommentLikeList = append(userCommentLikeList, userCommentLike)
	}

	return userCommentLikeList, nil
}

const userFindLikedCommentCountQuery = `SELECT comment_id , COUNT(*) 
	FROM user_comment_likes ucl
	GROUP BY comment_id  
`

func (s *User) FindLikedCommentCount(ctx context.Context) ([]*store.UserCommentLikeCountData, error) {
	userCommentLikeCountList := []*store.UserCommentLikeCountData{}

	rows, err := s.db.QueryContext(ctx, userFindLikedCommentCountQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		userCommentLikeCount := &store.UserCommentLikeCountData{}
		err := rows.Scan(
			&userCommentLikeCount.CommentID,
			&userCommentLikeCount.LikeCount,
		)
		if err != nil {
			return nil, err
		}
		userCommentLikeCountList = append(userCommentLikeCountList, userCommentLikeCount)
	}

	return userCommentLikeCountList, nil
}

const FindOneCommentByUserIDQuery = `SELECT uc.id, uc.comment
	FROM user_comments uc
	WHERE uc.user_id = $1
	LIMIT 1
`

func (s *User) FindOneCommentByUserID(ctx context.Context, userID string) (*store.UserCommentData, error) {
	userComment := &store.UserCommentData{}

	row := s.db.QueryRowContext(ctx, FindOneCommentByUserIDQuery, userID)

	err := row.Scan(
		&userComment.ID, &userComment.Comment,
	)
	if err != nil {
		return nil, err
	}

	return userComment, nil
}

const userUpdateCommentQuery = `UPDATE user_comments
	SET comment = $2, updated_at = $3
	WHERE id = $1
`

func (s *User) UpdateComment(ctx context.Context, userComment *store.UserCommentData) error {
	updateStmt, err := s.db.PrepareContext(ctx, userUpdateCommentQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		userComment.ID, userComment.Comment, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	userComment.UpdatedAt = sql.NullTime{Time: updatedAt, Valid: true}

	return nil

}

const insertUserRSVPQuery = `INSERT INTO
user_rsvps(
	id, user_id, people_count, created_at
) values(
	$1, $2, $3, $4
)
`

const userUpdateStatusQuery = `UPDATE users
	SET status = $2, updated_at = $3
	WHERE id = $1
`

func (s *User) InsertUserRSVP(ctx context.Context, userRSVP *store.UserRSVPData) error {
	insertStmt, err := s.db.PrepareContext(ctx, insertUserRSVPQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	userRSVPID := uuid.NewString()
	createdAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
		userRSVPID, userRSVP.UserID, userRSVP.PeopleCount, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	updatedAt := time.Now().UTC()
	updateStmt, err := s.db.PrepareContext(ctx, userUpdateStatusQuery)
	if err != nil {
		return err
	}
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		userRSVP.UserID, store.UserStatusRSVPProvided, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	userRSVP.ID = userRSVPID
	userRSVP.CreatedAt = createdAt

	return nil

}

const userRSVPUpdateByUserIDQuery = `UPDATE user_rsvps
	SET people_count = $2, updated_at = $3
	WHERE user_id = $1
`

func (s *User) UpdateRSVPByUserID(ctx context.Context, userRSVP *store.UserRSVPData) error {
	updateStmt, err := s.db.PrepareContext(ctx, userRSVPUpdateByUserIDQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		userRSVP.UserID, userRSVP.PeopleCount, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

const userFindAllWhatsAppNumberQuery = `SELECT u.wa_number 
FROM users u
ORDER by u.created_at asc
`

func (s *User) FindAllWhatsAppNumber(ctx context.Context) ([]string, error) {
	waNumberList := []string{}

	rows, err := s.db.QueryContext(ctx, userFindAllWhatsAppNumberQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var waNumber string

		err := rows.Scan(
			&waNumber,
		)
		if err != nil {
			return nil, err
		}
		waNumberList = append(waNumberList, waNumber)
	}

	return waNumberList, nil
}

const userRSVPAttendanceUpdateByUserIDQuery = `UPDATE user_rsvps
	SET is_attending = $2, updated_at = $3
	WHERE user_id = $1
`

func (s *User) UpdateRSVPAttendanceByUserID(ctx context.Context, userRSVP *store.UserRSVPData) error {
	updateStmt, err := s.db.PrepareContext(ctx, userRSVPAttendanceUpdateByUserIDQuery)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		userRSVP.UserID, userRSVP.IsAttending, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
