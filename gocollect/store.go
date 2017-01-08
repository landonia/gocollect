// Copyright 2017 Landonia Ltd. All rights reserved.

package gocollect

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/landonia/golog"
)

var (
	logger = golog.New("gollect.Store")

	// bucket names
	userBucketName            = []byte("users")
	userEventsBucketName      = []byte("userevents")
	emailToUserIDBucketName   = []byte("emailtouserid")
	phoneNoToUserIDBucketName = []byte("phonenotouserid")
)

// Store basically wraps the bolt DB and provides methods for retreiving and
// setting data
type Store struct {
	sync.Mutex
	open bool
	db   *bolt.DB
}

// Open the database specified by the path
func (s *Store) Open(path string) error {
	s.Lock()
	defer s.Unlock()
	var err error
	if s.db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 10 * time.Second}); err != nil {
		s.open = true
	}
	return err
}

// Init will initialise the Database by setting up the correct
// buckets if they do not already exist
func (s *Store) Init() error {

	// Create the user bucket
	err := s.db.Update(func(tx *bolt.Tx) error {

		// Used to store the user information based on a unique incrementing ID
		users, err := tx.CreateBucketIfNotExists(userBucketName)
		if err != nil {
			return err
		}

		// We need to create a user events bucket
		_, err = users.CreateBucketIfNotExists(userEventsBucketName)
		if err != nil {
			return err
		}

		// We will store a lookup that will lookup the id for the user using
		// the email address as the key
		_, err = users.CreateBucketIfNotExists(emailToUserIDBucketName)
		if err != nil {
			return err
		}

		// We will store a lookup that will lookup the id for the user using
		// the phone number as the key
		_, err = users.CreateBucketIfNotExists(phoneNoToUserIDBucketName)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// Size will return the size of the DB
func (s *Store) Size() int64 {
	var size int64
	s.db.View(func(tx *bolt.Tx) error {
		size = tx.Size()
		return nil
	})
	return size
}

// Backup will write the database to the writer
func (s *Store) Backup(w io.Writer) error {
	err := s.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
	return err
}

// Close the existing DB
func (s *Store) Close() error {
	s.Lock()
	defer s.Unlock()
	if !s.open {
		return fmt.Errorf("The DB connection is not open")
	}
	s.open = false
	return s.db.Close()
}

// GetUserIDUsingEmail will retrieve the ID for the user identified by the email
func (s *Store) GetUserIDUsingEmail(email string) (uint64, error) {

	// Read from the email to user index
	var id uint64
	err := s.db.View(func(tx *bolt.Tx) error {

		// Used to store the user information based on a unique incrementing ID
		usersBucket := tx.Bucket(userBucketName)
		if usersBucket != nil {

			// Attempt to lookup the email bucket
			emailBucket := usersBucket.Bucket(emailToUserIDBucketName)
			if emailBucket != nil {
				val := emailBucket.Get([]byte(email))
				if val != nil {
					id = idFromBytes(val)
					return nil
				}
			}
		}
		return fmt.Errorf("Could not find user ID for email: %s", email)
	})
	return id, err
}

// GetUserIDsMatchingFuzzyEmail will return all the IDs that match the
// email fuzzy search
func (s *Store) GetUserIDsMatchingFuzzyEmail(email string) ([]uint64, error) {

	// Read from the email to user index
	var ids []uint64
	err := s.db.View(func(tx *bolt.Tx) error {

		// Used to store the user information based on a unique incrementing ID
		usersBucket := tx.Bucket(userBucketName)
		if usersBucket != nil {

			// Attempt to lookup the email bucket
			emailBucket := usersBucket.Bucket(emailToUserIDBucketName)
			if emailBucket != nil {
				c := emailBucket.Cursor()

				// Seek based on using the email as the prefix
				prefix := []byte(email)
				for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
					logger.Debug("key=%s, value=%d\n", k, idFromBytes(v))
					ids = append(ids, idFromBytes(v))
				}
				return nil
			}
		}
		return fmt.Errorf("Could not find user ID for email: %s", email)
	})
	return ids, err
}

// GetUserIDUsingPhoneNo will retrieve the ID for the user identified by the phone number
func (s *Store) GetUserIDUsingPhoneNo(phoneNo string) (uint64, error) {

	// Read from the email to user index
	var id uint64
	err := s.db.View(func(tx *bolt.Tx) error {

		// Used to store the user information based on a unique incrementing ID
		usersBucket := tx.Bucket(userBucketName)
		if usersBucket != nil {

			// Attempt to lookup the phone number bucket
			phoneNoBucket := usersBucket.Bucket(phoneNoToUserIDBucketName)
			if phoneNoBucket != nil {
				val := phoneNoBucket.Get([]byte(phoneNo))
				if val != nil {
					id = idFromBytes(val)
					return nil
				}
			}
		}
		return fmt.Errorf("Could not find user ID for phone: %s", phoneNo)
	})
	return id, err
}

// GetUser will return the User identified
func (s *Store) GetUser(id uint64) (User, error) {

	// Read from the email to user index
	var user User
	err := s.db.View(func(tx *bolt.Tx) error {

		// Used to store the user information based on a unique incrementing ID
		usersBucket := tx.Bucket(userBucketName)
		if usersBucket != nil {
			val := usersBucket.Get(idToBytes(id))
			if val != nil {
				return json.Unmarshal(val, &user)
			}
		}
		return fmt.Errorf("Could not find user for ID: %d", id)
	})
	return user, err
}

// AddUser will add the user to the store if it does not currently exist
// and is valid
func (s *Store) AddUser(u User) (uint64, error) {

	// Ensure that the user is valid
	if !u.IsValid() {
		return 0, fmt.Errorf("User is not valid")
	}

	// Attempt to update the store with the user
	var id uint64
	err := s.db.Update(func(tx *bolt.Tx) error {

		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		usersBucket := tx.Bucket(userBucketName)

		// Generate ID for the user.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ = usersBucket.NextSequence()
		u.ID = id

		// Marshal user data into bytes.
		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		err = usersBucket.Put(idToBytes(u.ID), buf)
		if err != nil {
			return err
		}

		// Create a bucket for this user within the events bucket
		eventsBucket := usersBucket.Bucket(userEventsBucketName)
		if eventsBucket != nil {
			_, err = eventsBucket.CreateBucket(idToBytes(u.ID))
		}

		// Add the fields to the other indexes
		userEmailBucket := usersBucket.Bucket(emailToUserIDBucketName)
		if userEmailBucket != nil {
			err = userEmailBucket.Put([]byte(u.Email), idToBytes(u.ID))
			if err != nil {
				return err
			}
		}

		// If there is a phone number add this to the bucket index
		if u.PhoneNo != "" {
			userPhoneNoBucket := usersBucket.Bucket(phoneNoToUserIDBucketName)
			if userPhoneNoBucket != nil {
				err = userPhoneNoBucket.Put([]byte(u.PhoneNo), idToBytes(u.ID))
				if err != nil {
					return err
				}
			}
		}
		return err
	})
	return id, err
}

// idToBytes returns an 8-byte big endian representation of v.
func idToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// idFromBytes returns an uint64 using 8-byte big endian representation of bytes.
func idFromBytes(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
