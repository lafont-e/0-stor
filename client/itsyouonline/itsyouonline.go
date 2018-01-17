/*
 * Copyright (C) 2017-2018 GIG Technology NV and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package itsyouonline

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/itsyouonline/identityserver/clients/go/itsyouonline"
)

const (
	accessTokenURI = "https://itsyou.online/v1/oauth/access_token?response_type=id_token"
)

var (
	errNoPermission = errors.New("no permission")

	// ErrForbidden represents a forbidden action error
	ErrForbidden = errors.New("forbidden action")
)

// Config is used to create an IYO client.
type Config struct {
	Organization      string `yaml:"organization" json:"organization"`
	ApplicationID     string `yaml:"app_id" json:"app_id"`
	ApplicationSecret string `yaml:"app_secret" json:"app_secret"`
}

// Client defines itsyouonline client which is designed to help 0-stor user.
// It is not replacement for official itsyouonline client
type Client struct {
	cfg       Config
	iyoClient *itsyouonline.Itsyouonline
}

// NewClient creates new client
func NewClient(cfg Config) (*Client, error) {
	if cfg.Organization == "" {
		return nil, errors.New("IYO: organization not defined")
	}
	if cfg.ApplicationID == "" {
		return nil, errors.New("IYO: application ID not defined")
	}
	if cfg.ApplicationSecret == "" {
		return nil, errors.New("IYO: application Secret not defined")
	}
	return &Client{
		cfg:       cfg,
		iyoClient: itsyouonline.NewItsyouonline(),
	}, nil
}

// CreateJWT creates itsyouonline JWT token with these scopes:
// - org.namespace.read if perm.Read is true
// - org.namespace.write if perm.Write is true
// - org.namespace.delete if perm.Delete is true
func (c *Client) CreateJWT(namespace string, perm Permission) (string, error) {
	qp := map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     c.cfg.ApplicationID,
		"client_secret": c.cfg.ApplicationSecret,
		"validity":      "300", // 5 minutes, expressed in seconds
	}

	// build scopes query
	scopes := perm.Scopes(c.cfg.Organization, "0stor"+"."+namespace)
	if len(scopes) == 0 {
		return "", errNoPermission
	}
	qp["scope"] = strings.Join(scopes, ",")

	// create the request
	req, err := http.NewRequest("POST", accessTokenURI, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = buildQueryString(req, qp)

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get access token, response code = %v", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err

}

// Creates name as suborganization of org
func createSubOrganization(c *Client, org, suborg string) error {
	body := org + "." + suborg
	sub := itsyouonline.Organization{Globalid: body}

	_, resp, err := c.iyoClient.Organizations.CreateNewSubOrganization(org, sub, nil, nil)

	if err != nil {
		if resp.StatusCode == 409 {
			return fmt.Errorf("[Error] %s exists code=%v, err=%v", body, resp.StatusCode, err)
		}
		return fmt.Errorf("code=%v, err=%v", resp.StatusCode, err)
	}
	return nil
}

// CreateNamespace creates namespace as itsyouonline organization
// Verifies the full namespace path exists, and creates it if don't
// It also creates....
// - org.0stor.namespace.read
// - org.0stor.namespace.write
// - org.0stor.namespace.write
func (c *Client) CreateNamespace(namespace string) error {
	err := c.login()
	if err != nil {
		return err
	}

	// Verify c.cfg.Organization
	_, resp, err := c.iyoClient.Organizations.GetOrganization(c.cfg.Organization, nil, nil)

	if err != nil && resp.StatusCode != 403 {
		return fmt.Errorf("[Error] GetOrganization code=%v, err=%v", resp.StatusCode, err)
	}

	// Create c.cfg.Organization if non existent
	org := c.cfg.Organization
	if resp.StatusCode == 403 { // Forbiden, organization does not exist mainly
		organization := itsyouonline.Organization{Globalid: org}
		_, resp, err := c.iyoClient.Organizations.CreateNewOrganization(organization, nil, nil)
		// 200 => Org Created, 409 => Org Existed, other errors needs to be reported
		if err != nil && resp.StatusCode != 409 {
			return fmt.Errorf("[Error] CreateNewOrganization code=%v, err=%v", resp.StatusCode, err)
		}
		// Create c.cfg.Organization.0stor
		if err = createSubOrganization(c, org, "0stor"); err != nil {
			return err
		}
	}

	// Create c.cfg.Organization.0stor.namespace
	org += ".0stor"
	if err = createSubOrganization(c, org, namespace); err == nil {
		// Create c.cfg.Organization.0stor.namespace.read
		org += "." + namespace
		if err = createSubOrganization(c, org, "read"); err == nil {
			// Create c.cfg.Organization.0stor.namespace.write
			if err = createSubOrganization(c, org, "write"); err == nil {
				// Create c.cfg.Organization.0stor.namespace.delete
				err = createSubOrganization(c, org, "delete")
			}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// DeleteNamespace deletes the namespace sub organization and all of it's sub organizations
func (c *Client) DeleteNamespace(namespace string) error {
	err := c.login()
	if err != nil {
		return err
	}

	resp, err := c.iyoClient.Organizations.DeleteOrganization(
		c.createNamespaceID(namespace), nil, nil)
	if err != nil {
		return fmt.Errorf(
			"deleting namespace failed: IYO returned status %+v \nwith error message: %v",
			resp.Status, err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	return nil
}

// GivePermission give a user some permission on a namespace
func (c *Client) GivePermission(namespace, userID string, perm Permission) error {
	err := c.login()
	if err != nil {
		return err
	}

	var org string
	for _, perm := range perm.perms() {
		if perm == "admin" {
			org = c.createNamespaceID(namespace)
		} else {
			org = c.createNamespaceID(namespace) + "." + perm
		}
		user := itsyouonline.OrganizationsGlobalidMembersPostReqBody{Searchstring: userID}
		_, resp, err := c.iyoClient.Organizations.AddOrganizationMember(org, user, nil, nil)
		if err != nil {
			return fmt.Errorf("give member permission failed: code=%v, err=%v", resp.StatusCode, err)
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("give member permission failed: code=%v", resp.StatusCode)
		}
	}

	return nil
}

// RemovePermission remove some permission from a user on a namespace
func (c *Client) RemovePermission(namespace, userID string, perm Permission) error {
	err := c.login()
	if err != nil {
		return err
	}

	var org string
	for _, perm := range perm.perms() {
		if perm == "admin" {
			org = c.createNamespaceID(namespace)
		} else {
			org = c.createNamespaceID(namespace) + "." + perm
		}
		resp, err := c.iyoClient.Organizations.RemoveOrganizationMember(userID, org, nil, nil)
		if err != nil {
			return fmt.Errorf("removing permission failed: IYO returned status %+v \nwith error message: %v", resp.Status, err)
		}
	}

	return nil
}

// GetPermission retrieves the permission a user has for a namespace
// returns true for a right when user is member or invited to the namespace
func (c *Client) GetPermission(namespace, userID string) (Permission, error) {
	var (
		permission = Permission{}
		org        string
	)

	err := c.login()
	if err != nil {
		return permission, err
	}

	for _, perm := range []string{"read", "write", "delete", "admin"} {
		if perm == "admin" {
			org = c.createNamespaceID(namespace)
		} else {
			org = c.createNamespaceID(namespace) + "." + perm
		}

		invitations, resp, err := c.iyoClient.Organizations.GetInvitations(org, nil, nil)
		if err != nil {
			return permission, fmt.Errorf("Failed to retrieve user permission : %+v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return permission, fmt.Errorf("Failed to retrieve user permission : IYO returned status %+v", resp.Status)
		}

		members, resp, err := c.iyoClient.Organizations.GetOrganizationUsers(org, nil, nil)
		if err != nil {
			return permission, fmt.Errorf("Failed to retrieve user permission: %+v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return permission, fmt.Errorf("Failed to retrieve user permission : IYO returned status %+v", resp.Status)
		}

		switch perm {
		case "read":
			if hasPermission(userID, members.Users, invitations) {
				permission.Read = true
			}
		case "write":
			if hasPermission(userID, members.Users, invitations) {
				permission.Write = true
			}
		case "delete":
			if hasPermission(userID, members.Users, invitations) {
				permission.Delete = true
			}
		case "admin":
			if hasPermission(userID, members.Users, invitations) {
				permission.Admin = true
			}
		}
	}
	return permission, nil
}

func (c *Client) login() error {
	_, _, _, err := c.iyoClient.LoginWithClientCredentials(
		c.cfg.ApplicationID, c.cfg.ApplicationSecret)
	if err != nil {
		return fmt.Errorf("login failed:%v", err)
	}
	return nil
}

func (c *Client) createNamespaceID(namespace string) string {
	return c.cfg.Organization + "." + "0stor" + "." + namespace
}

func hasPermission(target string, members []itsyouonline.OrganizationUser, invitations []itsyouonline.JoinOrganizationInvitation) bool {
	return isMember(target, members) || isInvited(target, invitations)
}

func isMember(target string, list []itsyouonline.OrganizationUser) bool {
	for _, v := range list {
		if target == v.Username {
			return true
		}
	}
	return false
}

func isInvited(target string, invitations []itsyouonline.JoinOrganizationInvitation) bool {
	for _, invite := range invitations {
		if target == invite.User || target == invite.Emailaddress {
			return true
		}
	}
	return false
}

func buildQueryString(req *http.Request, qs map[string]interface{}) string {
	q := req.URL.Query()

	for k, v := range qs {
		q.Add(k, fmt.Sprintf("%v", v))
	}
	return q.Encode()
}
