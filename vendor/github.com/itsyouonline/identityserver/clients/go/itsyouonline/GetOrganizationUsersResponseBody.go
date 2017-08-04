package itsyouonline

import (
	"gopkg.in/validator.v2"
)

type GetOrganizationUsersResponseBody struct {
	Haseditpermissions bool               `json:"haseditpermissions" validate:"nonzero"`
	Users              []OrganizationUser `json:"users" validate:"nonzero"`
}

func (s GetOrganizationUsersResponseBody) Validate() error {

	return validator.Validate(s)
}
