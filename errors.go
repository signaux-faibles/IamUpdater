package main

import (
	"fmt"
)

type MisconfiguredUserError struct {
	user  User
	cause string
}

func (e MisconfiguredUserError) Error() string {
	return fmt.Sprintf("l'utilisateur %s est mal configuré : %s", e.user.email, e.cause)
}
