package handlers

import (
	"database/sql"
	"fmt"
	"harmony/backend/common"
	"time"

	"github.com/google/uuid"
)

type BufType string

const (
	TextType  BufType = "text"
	ImageType BufType = "image"
)

type User struct {
	Id    string
	Email string
}

type Buffer struct {
	Id     string
	UserId string
	Time   int64
	Ttl    int64
	Type   BufType
	Data   []byte
}

func GetBuffer(userid string) ([]byte, BufType, int64, error) {
	query := `
		SELECT data, type, ttl
		FROM buffer
		WHERE user_id = ?
		ORDER BY time DESC
		LIMIT 1`

	var data []byte
	var bufType string
	var ttl int64

	err := common.Db.QueryRow(query, userid).Scan(&data, &bufType, &ttl)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, TextType, 0, fmt.Errorf("no buffer found")
		}
		return nil, TextType, 0, err
	}

	if time.Now().Unix() > ttl {
		return nil, TextType, 0, fmt.Errorf("buffer expired")
	}

	bt := TextType
	if bufType == string(ImageType) {
		bt = ImageType
	}

	return data, bt, ttl, nil
}

func UpsertBuffer(userid string, data []byte, t BufType) (int64, error) {
	// Check if a buffer already exists for this user
	var existingId string
	query := `SELECT _id FROM buffer WHERE user_id = ? LIMIT 1`
	err := common.Db.QueryRow(query, userid).Scan(&existingId)

	var _id string
	if err == nil {
		_id = existingId
	} else if err == sql.ErrNoRows {
		_id = uuid.New().String()
	} else {
		return 0, err
	}

	ttl := time.Now().Add(common.Lifetime).Unix()
	currentTime := time.Now().Unix()

	// Use a transaction to ensure atomicity
	tx, err := common.Db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if existingId != "" {
		_, err = tx.Exec(`
			UPDATE buffer
			SET data = ?, time = ?, ttl = ?, type = ?
			WHERE _id = ?`,
			data, currentTime, ttl, string(t), _id)
	} else {
		_, err = tx.Exec(`
			INSERT INTO buffer (_id, user_id, time, ttl, type, data)
			VALUES (?, ?, ?, ?, ?, ?)`,
			_id, userid, currentTime, ttl, string(t), data)
	}

	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

func CreateOrGetUser(email string) (string, error) {
	var userId string
	err := common.Db.QueryRow(`SELECT _id FROM user WHERE email = ?`, email).Scan(&userId)

	if err == nil {
		return userId, nil
	} else if err != sql.ErrNoRows {
		return "", err
	}

	uid := uuid.New().String()

	tx, err := common.Db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check again in transaction to handle race conditions
	err = tx.QueryRow(`SELECT _id FROM user WHERE email = ?`, email).Scan(&userId)
	if err == nil {
		// Another process created the user, return that ID
		tx.Commit()
		return userId, nil
	} else if err != sql.ErrNoRows {
		return "", err
	}

	// Create new user
	_, err = tx.Exec(`INSERT INTO user (_id, email) VALUES (?, ?)`, uid, email)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return uid, nil
}
