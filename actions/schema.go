// This Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"encoding/gob"
	"math/rand"
	"time"

	"github.com/Ronmi/yugi/config"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	types := []any{
		&Org{},
		&User{},
		&Schedule{},
		&Appointment{},
		&Receipt{},
		&Log{},
	}

	for _, t := range types {
		gob.Register(t)
	}
	return db.AutoMigrate(types...)

}

// Org 是一個罷免團隊
//
// - Name: 團隊名稱 (斷健行動)
// - Area: 罷免區域 (中五、北區、北屯區)
// - Target: 罷免目標 (黃健豪)
// - UseSubmission: 是否啟用交件功能
type Org struct {
	ID            int64
	Name          string   `gorm:"unique"`
	Area          []string `gorm:"serializer:json"`
	Target        string
	UseSubmission bool
}

// for template validation
func FakeOrg() *Org {
	return &Org{
		ID:            int64(rand.Intn(100) + 1),
		Name:          config.RandName(),
		Area:          []string{config.RandName()},
		Target:        config.RandName(),
		UseSubmission: true,
	}
}

type OauthProvider string

const (
	Google    OauthProvider = "google"
	Facebook  OauthProvider = "facebook"
	Instagram OauthProvider = "instagram"
	Twitter   OauthProvider = "twitter"
	Line      OauthProvider = "line"
	Telegram  OauthProvider = "telegram"
)

type Role string

const (
	Manager   Role = "manager"   // 罷團幹部
	Volunteer Role = "volunteer" // 罷免志工
	Novice    Role = "novice"    // 新手志工
	Member    Role = "member"    // 一般民眾
)

// 使用者帳號
//
// - Name: 名稱匿名化成 "隊員#甲乙丙丁" 這種格式，這裡存的是 "甲乙丙丁"，隨機產生
// - Role: 使用者角色，決定權限
type User struct {
	ID            int64
	OauthID       string        `gorm:"uniqueIndex:idx_oauth_provider_oauth_id"`
	OauthProvider OauthProvider `gorm:"uniqueIndex:idx_oauth_provider_oauth_id"`
	Name          string        `gorm:"unique"`
	Note          string        // 幹部志工都能看見/僅幹部能修改
	Secret        string        // 僅幹部能看見修改
	Role          Role
	OrgID         *int64
	TOTPSecret    string

	Org *Org `gorm:"foreignKey:OrgID"`
}

// for template validation
func (u *User) SetRole(r Role) *User {
	u.Role = r
	return u
}

// for template validation
func (u *User) SetOrg(o *Org) *User {
	if o == nil {
		u.OrgID = nil
		u.Org = nil
		return u
	}

	u.OrgID = &o.ID
	u.Org = o
	return u
}

// for template validation
func (u *User) SetTOTPSecret(secret string) *User {
	k, _ := config.CreateOTPKey(u.Name)
	u.TOTPSecret = k.Secret()
	return u
}

// for template validation
func FakeUser() *User {
	return &User{
		ID:            int64(rand.Intn(100) + 1),
		OauthID:       config.RandName(),
		OauthProvider: Telegram,
		Name:          config.RandName(),
		Role:          Member,
		OrgID:         nil,
		TOTPSecret:    "",
	}
}

// Schedule 是一個可預約的項目，User 是志工
//
// - BeginAt: 開始時間
// - EndAt: 結束時間
// - Area: 說明此志工有可能前往收件的區域
// - Disabled: 是否停用，停用後不會出現在可掛號的清單中
type Schedule struct {
	ID       int64
	UserID   int64
	BeginAt  time.Time
	EndAt    time.Time
	Area     string
	Disabled bool

	User *User `gorm:"foreignKey:UserID"`
}

// for template validation
func (s *Schedule) SetDisabled(d bool) *Schedule {
	s.Disabled = d
	return s
}

// for template validation
func FakeSchedule(u *User) *Schedule {
	n := time.Duration(-rand.Intn(100)) * time.Second
	return &Schedule{
		ID:       int64(rand.Intn(100) + 1),
		UserID:   u.ID,
		BeginAt:  time.Now().Add(n),
		EndAt:    time.Now().Add(n + time.Hour),
		Area:     config.RandName(),
		Disabled: false,
		User:     u,
	}
}

type AppointmentStatus int

// 狀態分為已完成和未完成兩類，Cancelled, Completed, Missed 是已完成的狀態
const (
	Pending    AppointmentStatus = iota // 預設狀態，等待志工確認
	Contacting                          // 聯絡中
	Confirmed                           // 預約成功
	NotMatched                          // 時間無法配合，需幹部介入轉給其他志工
	Cancelled                           // 取消預約
	Completed                           // 已取件
	Missed                              // 連署人未出現
	InvalidStatus
)

var AvaialbeAppointmentStatus = map[string]AppointmentStatus{
	"pending":    Pending,
	"contacting": Contacting,
	"confirmed":  Confirmed,
	"notMatched": NotMatched,
	"cancelled":  Cancelled,
	"completed":  Completed,
	"missed":     Missed,
}

// ContactMethod 是民眾提供 聯絡方式及指示
type ContactMethod struct {
	Name   string // 如何稱呼 (黃先生)
	Prefer string // 偏好的聯絡方式 (周一至五白天)
	Phone  string // 電話號碼
}

// for template validation
func FakeContactMethod() *ContactMethod {
	return &ContactMethod{
		Name:   config.RandName(),
		Prefer: config.RandName(),
		Phone:  config.RandName(),
	}
}

// Appointment 是一個掛號/預約，User 是一般民眾
//
// - Name: 如何稱呼 (黃先生) (民眾填寫)
// - UserNote: 寫給民眾看的備註
// - VolunteerNote: 寫給志工或幹部看的備註
// - ContactAt: 標記為聯絡中的時間
// - FinishAt: 變更為 Completed/Missed/Cancelled 的時間 (之後不能再變更)
type Appointment struct {
	ID            int64
	UserID        int64
	ScheduleID    int64
	RegisterAt    time.Time
	ContactAt     *time.Time
	ContactMethod *ContactMethod `gorm:"serializer:json"`
	UserNote      string
	VolunteerNote string
	FinishAt      *time.Time
	Status        AppointmentStatus

	User     *User     `gorm:"foreignKey:UserID"`
	Schedule *Schedule `gorm:"foreignKey:ScheduleID"`
	Receipt  *Receipt  `gorm:"foreignKey:AppointmentID"`
}

// for template validation
func (s *Appointment) SetReceipt(r *Receipt) *Appointment {
	s.Receipt = r
	return s
}

// for template validation
func (s *Appointment) SetStatus(status AppointmentStatus) *Appointment {
	s.Status = status
	return s
}

// for template validation
func FakeAppointment(u *User, s *Schedule) *Appointment {
	return &Appointment{
		ID:            int64(rand.Intn(100) + 1),
		UserID:        u.ID,
		ScheduleID:    s.ID,
		RegisterAt:    time.Now(),
		ContactAt:     nil,
		ContactMethod: FakeContactMethod(),
		UserNote:      config.RandName(),
		VolunteerNote: config.RandName(),
		FinishAt:      nil,
		Status:        Pending,
		User:          u,
		Schedule:      s,
	}
}

// Receipt 是簽收記錄
//
// - Secret: 民眾填寫的簽收備註 (密語)
// - Note: 志工填寫的簽收備註
// - Receives: 收到的連署書 (區域名稱 -> 數量)
// - SubmitAt: 交件給罷團的時間
// - Manager: 負責收下這次交件的罷團幹部
type Receipt struct {
	ID              int64
	AppointmentID   int64 `gorm:"unique"`
	CreatedAt       time.Time
	CreatedByUserID int64
	SignAt          *time.Time
	Secret          string
	Note            string
	Receives        map[string]int `gorm:"serializer:json"`
	SubmitAt        *time.Time
	ManagerID       int64

	Appointment   *Appointment `gorm:"foreignKey:AppointmentID"`
	CreatedByUser *User        `gorm:"foreignKey:CreatedByUserID"`
	Manager       *User        `gorm:"foreignKey:ManagerID"`
}

// for template validation
func (r *Receipt) Receive(area string, num int) *Receipt {
	if r.Receives == nil {
		r.Receives = map[string]int{}
	}
	r.Receives[area] = num
	return r
}

// for template validation
func FakeReceipt(a *Appointment) *Receipt {
	now := time.Now()
	return &Receipt{
		ID:              int64(rand.Intn(100) + 1),
		AppointmentID:   a.ID,
		CreatedAt:       time.Now(),
		CreatedByUserID: a.UserID,
		SignAt:          &now,
		Secret:          config.RandName(),
		Note:            config.RandName(),
		Receives:        map[string]int{config.RandName(): rand.Intn(100)},
		Appointment:     a,
		CreatedByUser:   a.User,
	}
}

// Action 是操作類型
type Action string

const (
	RegisterAccount   Action = "register_account"
	UserLogin         Action = "user_login"
	GrantRole         Action = "grant_role"
	RevokeRole        Action = "revoke_role"
	Enable2FA         Action = "enable_2fa"
	EditMemberNote    Action = "edit_member_note"
	EditMemberSecret  Action = "edit_member_secret"
	EditMemberOTP     Action = "edit_member_otp"
	NewSchedule       Action = "new_schedule"
	DisableSchedule   Action = "disable_schedule"
	NewAppointment    Action = "new_appointment"
	CancelAppointment Action = "cancel_appointment"
	EditAppointment   Action = "edit_appointment"
	PairAppointment   Action = "pair_appointment"
	CreateReceipt     Action = "create_receipt"
	SignReceipt       Action = "sign_receipt"
)

// Log 是操作記錄
type Log struct {
	ID          int64
	Action      Action
	UserID      int64
	Description string
	CreatedAt   time.Time

	User *User `gorm:"foreignKey:UserID"`
}

// for template validation
func FakeLog(u *User) *Log {
	return &Log{
		ID:          int64(rand.Intn(100) + 1),
		Action:      GrantRole,
		UserID:      u.ID,
		Description: config.RandName(),
		CreatedAt:   time.Now(),
		User:        u,
	}
}
