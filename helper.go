package main

import (
	"mcp-server/database"

	"gorm.io/gorm"
)

const (
	UserRoleCeo             = "ceo"
	UserRoleAdmin           = "admin"
	UserRoleEmployee        = "employee"
	UserRoleTeamLeader      = "team_leader"
	UserRoleTeamleaderandHR = "Team Leader & HR"
)

func GetUserLeader(db *gorm.DB, user database.User, role database.Role) (*database.User, error) {
	if role.Name == UserRoleCeo {
		return &user, nil
	}

	if user.TeamID > 0 && role.Name == UserRoleEmployee {
		var teamLeader *database.User
		if err := db.First(&teamLeader, user.Team.LeaderID).Error; err != nil {
			return nil, err
		}
		return teamLeader, nil
	} else if role.Name == UserRoleTeamLeader {
		var ceoUser *database.User
		if err := db.Joins("left join pmt_roles on pmt_users.role_id = pmt_roles.id").Where("pmt_roles.name = ?", UserRoleCeo).First(&ceoUser).Error; err != nil {
			return nil, err
		}
		return ceoUser, nil
	}
	return nil, nil
}
