// Copyright 2017 Landonia Ltd. All rights reserved.

package gocollect

import "regexp"

var (
	// Regular expression that matches 99% of the email addresses in use today.
	validEmail = regexp.MustCompile(`(?i)[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,6}`)
)

// User holds the information for a particular user
// the unique element is the email address
type User struct {
	ID      uint64 `json:"id"`    // The unique ID for this user record
	Email   string `json:"email"` // The user email address
	PhoneNo string `json:"phone"` // The user phone number
}

// IsValid will return true if the user data is valid
func (u *User) IsValid() bool {
	return IsEmailValid(u.Email)
}

// IsEmailValid will determine if the email address is valid
func IsEmailValid(email string) bool {
	return validEmail.MatchString(email)
}
