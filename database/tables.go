package database

import (
	"time"

	"gorm.io/gorm"
)

type (
	Base struct {
		ID        uint           `gorm:"primarykey" json:"id"`                         //
		CreatedAt time.Time      `gorm:"column:created_at;not null" json:"created_at"` //
		UpdatedAt time.Time      `gorm:"column:updated_at;not null" json:"updated_at"` //
		DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                               //
	}

	Team struct {
		Base
		Name          string  `gorm:"column:name;unique;not null" json:"name"`         // Team нэр
		Description   *string `gorm:"column:description" json:"description"`           //
		IsActive      bool    `gorm:"column:is_active;default:false" json:"is_active"` // Идэвхтэй эсэх
		Color         string  `gorm:"column:color" json:"color"`                       // UI дээр харагдах өнгө
		LeaderID      uint    `gorm:"column:leader_id;not null" json:"leader_id"`      // Багын удирдагчийн ID
		Leader        *User   `gorm:"foreignKey:LeaderID" json:"leader"`               //
		Members       []*User `gorm:"foreignKey:TeamID" json:"members"`                // Багын гишүүд
		CreatedUserID *uint   `gorm:"column:created_user_id" json:"created_user_id"`   // Үүсгэсэн ажилтан
		CreatedUser   *User   `gorm:"foreignKey:CreatedUserID" json:"created_user"`    // Үүсгэсэн ажилтан
		ModifiedByID  *uint   `gorm:"column:modified_by_id" json:"modified_by_id"`     // Засварласан ажилтан
		ModifiedBy    *User   `gorm:"foreignKey:ModifiedByID" json:"modified_by"`      // Засварласан ажилтан
		GroupMail     string  `gorm:"column:group_mail" json:"group_mail"`
	}

	User struct {
		Base
		IsActive           bool      `gorm:"column:is_active;default:false" json:"is_active"` // Идэвхтэй эсэх
		Email              string    `gorm:"column:email" json:"email"`                       //
		FirstName          string    `gorm:"column:first_name;" json:"first_name"`            // Өөрийн нэр
		LastName           string    `gorm:"column:last_name;" json:"last_name"`              // Овог нэр
		NickName           string    `gorm:"column:nick_name;" json:"nick_name"`              // Хоч нэр
		ProfileID          *uint     `gorm:"column:profile_id" json:"profile_id"`             //
		Profile            *File     `gorm:"foreignKey:ProfileID" json:"profile"`             //
		CoverID            *uint     `gorm:"column:cover_id" json:"cover_id"`                 //
		Cover              *File     `gorm:"foreignKey:CoverID" json:"cover"`                 //
		Position           string    `gorm:"column:position" json:"position"`                 // Албан тушаал
		Bio                string    `gorm:"column:bio" json:"bio"`                           // Тайлбар
		PhoneNumber        string    `gorm:"column:phone_number" json:"phone_number"`         // Утасны дугаар
		IsFullTime         bool      `gorm:"column:is_full_time" json:"is_full_time"`         // Бүтэн цагийн ажилтан эсэх
		IsTemprary         bool      `gorm:"column:is_temprary" json:"is_temprary"`           // Жинхлэгдсэн эсэх
		Birthday           time.Time `gorm:"column:birthday;" json:"birthday"`                // Төрсөн өдөр
		Password           string    `gorm:"column:password;not null" json:"-"`               //
		TeamID             uint      `gorm:"column:team_id" json:"team_id"`                   //
		Team               *Team     `gorm:"foreignKey:TeamID" json:"team"`                   //
		LastLoginDate      time.Time `gorm:"column:last_login_date" json:"last_login_date"`   // Сүүлд нэвтэрсэн огноо
		EmploymentDate     time.Time `gorm:"column:employment_date" json:"employment_date"`
		RoleID             uint      `gorm:"role_id" json:"role_id"`
		Role               *Role     `gorm:"foreignKey:RoleID" json:"role"`
		DateOfNonTemprary  time.Time `gorm:"column:date_of_non_temprary"` //Ажилд орсон огноо
		TelegramChannel    string    `gorm:"column:telegram_channel" json:"telegram_channel"`
		RegistrationNumber string    `gorm:"column:registration_number" json:"registration_number"` // РД
		EncryptedRD        string    `gorm:"column:encrypted_rd" json:"-"`                          // For retrieval
		ResignationDate    time.Time `gorm:"column:resignation_date" json:"resignation_date"`
		LastNameMN         string    `gorm:"column:last_name_mn" json:"last_name_mn"`
		FirstNameMN        string    `gorm:"column:first_name_mn" json:"first_name_mn"`
		Gender             string    `gorm:"column:gender" json:"gender"`
		Zodiac             string    `gorm:"column:zodiac" json:"zodiac"`
		Interests          string    `gorm:"column:interests" json:"interests"`
	}

	Role struct {
		Base
		Name string `gorm:"column:name" json:"name"`
		// Permissions []*Permission `gorm:"many2many:role_permissions" json:"permissions"`
		ParentID    uint   `gorm:"parent_id" json:"parent_id"`
		Parent      *Role  `gorm:"parent"  json:"parent"`
		DisplayName string `gorm:"-" json:"display_name"`
		// Menus       []*Menu       `gorm:"many2many:role_menus" json:"menus"`
	}

	File struct {
		Base
		OriginalName  string `gorm:"column:original_name; not null" json:"original_name"` // Файлын жинхэнэ нэр
		FileName      string `gorm:"column:file_name; not null" json:"file_name"`         // Файлыг хадгалах үед өөрчилсөн нэр
		Extention     string `gorm:"column:extention" json:"extention"`                   // Файлын төрөл
		PhysicalPath  string `gorm:"column:physical_path; not null" json:"physical_path"` // Файлын зам
		FileSize      int64  `gorm:"column:file_size; not null" json:"file_size"`         // Файлын хэмжээ
		CreatedUserID *uint  `gorm:"column:created_user_id" json:"created_user_id"`       //
		CreatedUser   *User  `gorm:"foreignKey:CreatedUserID" json:"created_user"`        //
	}

	Absence struct {
		Base
		IsVacationCalculation bool    `gorm:"column:is_vacation_calculation;default:false" json:"is_vacation_calculation"`
		EmployeeID            uint    `gorm:"column:employee_id" json:"employee_id"`         // EmployeeID
		Employee              *User   `gorm:"foreignkey:EmployeeID" json:"employee"`         // Employee
		Reason                string  `gorm:"column:reason" json:"reason"`                   // Reason
		InActiveHours         float64 `gorm:"column:in_active_hours" json:"in_active_hours"` //  In Active hours
		Description           string  `gorm:"column:description" json:"description"`         // Description
		IntervalID            uint          `gorm:"column:interval_id" json:"interval_id"`         // Interval ID
		Interval              *TimeInterval `gorm:"foreignKey:IntervalID" json:"interval"`         // Interval
		CreatedUserID uint      `gorm:"column:created_user_id" json:"created_user_Id"` // Created User ID
		CreatedUser   *User     `gorm:"foreignKey:CreatedUserID" json:"created_user"`  //  Created User
		RemainHours   float64   `gorm:"column:remain_hours" json:"remain_hours"`       // Remain Hours
		StartDate     time.Time `gorm:"column:start_date" json:"start_date"`           // Start Date
		Status        string    `gorm:"column:status" json:"status"`
		LeaderID      uint      `gorm:"column:leader_id" json:"leader_id"`
		Leader        *User     `gorm:"foreignKey:LeaderID" json:"leader"`
	}

	TimeInterval struct {
		Base
		Name      string    `gorm:"column:name;not null" json:"name"`                                               //
		BeginDate time.Time `gorm:"column:begin_date;not null" json:"begin_date"`                                   // Мөчлөгын эхлэх огноо
		EndDate   time.Time `gorm:"column:end_date;not null" json:"end_date"`                                       // Мөчлөгын дуусах огноо
		ItOver    bool      `grom:"it_over" json:"it_over"`                                                         // Дууссан эсэх
		// Tasks     []*Task   `gorm:"foreignKey:ModuleID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"tasks"` // Ажилбарын даалгаварууд
	}

)
