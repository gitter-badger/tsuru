package gitosis

import (
	"errors"
	"fmt"
	ini "github.com/kless/goconfig/config"
	"github.com/timeredbull/tsuru/config"
	"github.com/timeredbull/tsuru/log"
	"path"
	"strings"
)

// Add a new project to gitosis.conf.
func AddProject(group, project string) error {
	c, err := getConfig()
	if err != nil {
		return err
	}
	section := fmt.Sprintf("group %s", group)
	c.AddOption(section, "writable", project)
	return nil
}

// Add a new group to gitosis.conf. Also commit and push changes.
func AddGroup(name string) error {
	c, err := getConfig()
	if err != nil {
		return err
	}
	sName := fmt.Sprintf("group %s", name)
	ok := c.AddSection(sName)
	if !ok {
		errStr := fmt.Sprintf(`Could not add section "group %s" in gitosis.conf, section already exists!`, name)
		return errors.New(errStr)
	}
	commitMsg := fmt.Sprintf("Defining gitosis group for group %s", name)
	return writeCommitPush(c, commitMsg)
}

// Removes a group section and all it's options.
func RemoveGroup(group string) error {
	c, err := getConfig()
	if err != nil {
		return err
	}
	gName := fmt.Sprintf("group %s", group)
	ok := c.RemoveSection(gName)
	if !ok {
		return errors.New("Section does not exists")
	}
	commitMsg := fmt.Sprintf("Removing group %s from gitosis.conf", group)
	return writeCommitPush(c, commitMsg)
}

// addMember adds a member to the given group.
// member parameter should be the same as the key name in keydir dir.
func addMember(group, member string) error {
	c, err := getConfig()
	if err != nil {
		return err
	}
	section := fmt.Sprintf("group %s", group)
	if !c.HasSection(section) {
		return errors.New("Group not found")
	}
	var members []string
	if strMembers, err := c.String(section, "members"); err == nil {
		members = strings.Split(strMembers, " ")
	}
	if checkPresenceOfString(members, member) {
		return errors.New("This user is already member of this group")
	}
	members = append(members, member)
	c.AddOption(section, "members", strings.Join(members, " "))
	commitMsg := fmt.Sprintf("Adding member %s to group %s", member, group)
	return writeCommitPush(c, commitMsg)
}

// removeMember removes a member from the given group.
func removeMember(group, member string) error {
	c, err := getConfig()
	if err != nil {
		return err
	}
	section := fmt.Sprintf("group %s", group)
	if !c.HasSection(section) {
		return errors.New("Group not found")
	}
	strMembers, err := c.String(section, "members")
	if err != nil {
		return errors.New("This group does not have any members")
	}
	members := strings.Split(strMembers, " ")
	index := find(members, member)
	if index < 0 {
		return errors.New("This group does not have this member")
	}
	last := len(members) - 1
	members[index] = members[last]
	members = members[:last]
	if len(members) > 0 {
		c.AddOption(section, "members", strings.Join(members, " "))
	} else {
		c.RemoveOption(section, "members")
	}
	commitMsg := fmt.Sprintf("Removing member %s from group %s", member, group)
	return writeCommitPush(c, commitMsg)
}

func ConfPath() (p string, err error) {
	p = ""
	repoPath, err := config.GetString("git:gitosis-repo")
	if err != nil {
		log.Print(err)
		return
	}
	p = path.Join(repoPath, "gitosis.conf")
	return
}

func find(strs []string, str string) int {
	for i, s := range strs {
		if str == s {
			return i
		}
	}
	return -1
}

func checkPresenceOfString(strs []string, str string) bool {
	return find(strs, str) > -1
}

func getConfig() (*ini.Config, error) {
	confPath, err := ConfPath()
	if err != nil {
		return nil, err
	}
	return ini.ReadDefault(confPath)
}
