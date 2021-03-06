package main

import (
	"bufio"
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/xlab/treeprint"
	"gorm.io/gorm"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// For role granting.
// Implements database/models/HasRole interface.
type dummyHasRole struct {
	ID   uint
	name string
}

func (t *dummyHasRole) GetID() uint {
	return t.ID
}
func (t *dummyHasRole) TypeName() string {
	return t.name
}

func permission() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			log.Fatal("Edit permission failed.")
		}
	}()

	readConfig()
	initGorm()
	initLog()

	reader := bufio.NewReaderSize(os.Stdin, 0)
	if len(args) == 1 {
		quit := false
		log.Debug(`Entering interactive mode, enter "help" for help.`)
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			<-s
			os.Exit(0)
		}()
		for !quit {
			_, err := color.New(color.Bold).Print("EduOJ Permission> ")
			if err != nil {
				log.Error(errors.Wrap(err, "fail to print"))
				log.Fatal("Edit permission failed.")
				return
			}
			input, err := reader.ReadString('\n')
			if err != nil {
				log.Error(errors.Wrap(err, "fail to print"))
				return
			}
			if err == io.EOF {
				return
			}
			args = strings.Split(input[:len(input)-1], " ")
			quit = doPermission(args)
		}
	} else {
		doPermission(args[1:])
	}
}

func doPermission(args []string) (end bool) {
	var err error
	switch args[0] {
	case "help", "h":
		fmt.Println(`EduOJ Permission

Usage:
  One-line execution: $ EduOJ (permission|perm) (command) <args>...

  Enter interactive mode: $ EduOJ (permission|perm)
  Command format in interactive mode:  (command) <args>...

commands:
  (help|h)
  (list-roles|lr) [<role_id|role_name>]
  (create-role|cr) <name> [<target>]
  (grant-role|gr) <user_id|username> <role_id|role_name> [<target_id>]
  (delete-role|dr) <role_id|role_name>
  (add-permission|ap) <role_id|role_name> <permission>
  (quit|q)

Note:
  When the search value matches the name and ID at the same time, the system
  always selects the object that matches the ID.`)
	case "create-role", "cr":
		// (create-role|cr) <name> [<target>]
		err = validateArgumentsCount(len(args), 2, 3)
		if err != nil {
			log.Error(err)
			break
		}
		r := models.Role{
			Name: args[1],
		}
		if len(args) == 3 {
			r.Target = &args[2]
		}
		err = base.DB.Create(&r).Error
	case "list-roles", "lr":
		// (list-roles|lr) [<role_id|role_name>]
		err = validateArgumentsCount(len(args), 1, 2)
		tree := treeprint.New()
		tree.SetValue("Roles")
		if len(args) == 1 {
			var roles []models.Role
			err = base.DB.Set("gorm:auto_preload", true).Find(&roles).Error
			if err != nil {
				log.Error(err)
				break
			}
			for _, role := range roles {
				listRole(tree, &role)
			}
		} else {
			var role *models.Role
			role, err = findRole(args[1])
			if err != nil {
				log.Error(err)
				break
			}
			listRole(tree, role)
		}
		fmt.Println(tree.String())

	case "grant-role", "gr":
		// (grant-role|gr) <user_id|username> <role_id|role_name> [<target_id>]
		err = validateArgumentsCount(len(args), 3, 4)
		if err != nil {
			log.Error(err)
			break
		}
		var user *models.User
		user, err = utils.FindUser(args[1])
		if err != nil {
			err = errors.Wrap(err, "find user")
			log.Error(err)
			break
		}
		var role *models.Role
		role, err = findRole(args[2])
		if err != nil {
			log.Error(err)
			break
		}
		if len(args) == 3 {
			user.GrantRole(role.Name)
		} else {
			var targetId uint64
			targetId, err = strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				log.Error(err)
				break
			}
			target := dummyHasRole{
				ID:   uint(targetId),
				name: *role.Target,
			}
			user.GrantRole(role.Name, &target)
		}
	case "delete-role", "dr":
		// (delete-role|dr) <role_id|role_name>
		err = validateArgumentsCount(len(args), 2, 2)
		var role *models.Role
		role, err := findRole(args[1])
		if err != nil {
			log.Error(err)
			break
		}
		err = base.DB.Delete(&models.Permission{}, "role_id = ?", role.ID).Error
		if err != nil {
			log.Error(err)
			break
		}
		err = base.DB.Delete(&role).Error
		if err != nil {
			log.Error(err)
			break
		}
	case "add-permission", "ap":
		// (add-permission|ap) <role_id|role_name> <permission>
		err = validateArgumentsCount(len(args), 3, 3)
		if err != nil {
			log.Error(err)
			break
		}
		var role *models.Role
		role, err = findRole(args[1])
		if err != nil {
			log.Error(err)
			break
		}
		role.AddPermission(args[2])
	case "quit", "q":
		return true
	default:
		log.Debug("Unknown operation \"" + args[0] + "\".")
	}
	return false
}

func findRole(id string) (*models.Role, error) {
	role := models.Role{}
	err := base.DB.Set("gorm:auto_preload", true).Where("id = ?", id).First(&role).Error
	if err != nil {
		err = base.DB.Set("gorm:auto_preload", true).Where("name = ?", id).First(&role).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("role record not found")
			} else {
				return nil, errors.Wrap(err, "could not query role")
			}
		}
	}
	return &role, nil
}

func validateArgumentsCount(count int, min int, max int) (err error) {
	if count < min {
		err = errors.New("Too few command line parameters")
	} else if count > max {
		err = errors.New("Too many command line parameters")
	}
	return
}

func listRole(root treeprint.Tree, role *models.Role) {
	roleString := role.Name
	if role.Target != nil {
		roleString += "(" + color.YellowString(*role.Target) + ")"
	}
	roleNode := root.AddBranch(roleString + "[" + color.MagentaString("%d", role.ID) + "]")

	for _, perm := range role.Permissions {
		roleNode.AddNode(color.GreenString(perm.Name) + "[" + color.MagentaString("%d", perm.ID) + "]")
	}
	return
}
