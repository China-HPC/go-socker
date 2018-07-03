package user

import (
	"os/user"
	"strconv"
	"syscall"
)

// User represents a user account.
type User struct {
	UID   int
	GID   int
	Name  string
	Home  string
	Shell string
}

// Group represents a group account.
type Group struct {
	GID   int
	Name  string
	Users []string
}

// UserCred contains user's credential and user info.
type UserCred struct {
	User *user.User
	Cred *syscall.Credential
}

// GetUserCred returns a credential of specified user.
func GetUserCred(username string) (*UserCred, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	return getUserCredByUID(u)
}

// GetUserCredByUID returns a credential of specified user.
func GetUserCredByUID(uid string) (*UserCred, error) {
	u, err := user.LookupId(uid)
	if err != nil {
		return nil, err
	}
	return getUserCredByUID(u)
}

func getUserCredByUID(u *user.User) (*UserCred, error) {
	unitUID, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	unitGID, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gids, err := u.GroupIds()
	if err != nil {
		return nil, err
	}
	var groups []uint32
	for _, gid := range gids {
		uintGid, err := strconv.ParseUint(gid, 10, 32)
		if err != nil {
			return nil, err
		}
		groups = append(groups, uint32(uintGid))
	}
	return &UserCred{
		User: u,
		Cred: &syscall.Credential{
			Uid:    uint32(unitUID),
			Gid:    uint32(unitGID),
			Groups: groups,
		},
	}, nil
}
