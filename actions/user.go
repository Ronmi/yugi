// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"errors"
	"fmt"

	"github.com/Ronmi/yugi/config"
	"gorm.io/gorm"
)

func createUser(tx *gorm.DB, id string, provider OauthProvider) (*User, error) {
	u := User{
		OauthID:       id,
		OauthProvider: provider,
		Role:          Member,
	}
	cnt := 0
	err := gorm.ErrDuplicatedKey
	for errors.Is(err, gorm.ErrDuplicatedKey) && cnt < 10 {
		u.Name = config.RandName()
		err = tx.Create(&u).Error
		cnt++
	}

	err = LogByUID(tx, u.ID, RegisterAccount, fmt.Sprintf(
		"您使用 %s 註冊成功",
		provider,
	))
	if err != nil {
		return nil, err
	}

	return &u, err
}

func FindUser(tx *gorm.DB, name string) (*User, error) {
	var u User
	err := tx.
		Where("name = ?", name).
		Preload("Org").
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUserByID(tx *gorm.DB, id string) (*User, error) {
	var u User
	err := tx.
		Where("id = ?", id).
		Preload("Org").
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Login 透過 OAuth ID 登入，不存在則直接新增
func Login(tx *gorm.DB, id string, provider OauthProvider) (ret *User, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var u User
		err = tx.
			Where("oauth_id = ? AND oauth_provider = ?", id, provider).
			Preload("Org").
			First(&u).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				ret, err = createUser(tx, id, provider)
				return
			}
		}

		msg := "您從 %s 登入成功"
		if u.TOTPSecret != "" {
			msg = "您從 %s 嘗試登入 (需驗證 2FA)"
		}
		err = LogByUID(tx, u.ID, UserLogin, fmt.Sprintf(
			msg,
			provider,
		))
		if err == nil {
			ret = &u
		}

		return
	})
	return
}

func SaveTOTP(tx *gorm.DB, u *User, secret string) error {
	return tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(u).Update("totp_secret", secret).Error; err != nil {
			return err
		}

		return LogByUID(tx, u.ID, Enable2FA, "")
	})
}

func ListMembers(tx *gorm.DB, orgID int64) ([]*User, error) {
	var users []*User
	err := tx.
		Where("org_id = ?", orgID).
		Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return users, nil
}

func OrgGrantRole(tx *gorm.DB, mgr *User, uname string, role Role) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var victim User
		err = tx.Where("name = ?", uname).First(&victim).Error
		if err != nil {
			return
		}
		err = tx.Model(&victim).
			Updates(map[string]any{
				"role":   role,
				"org_id": mgr.OrgID,
			}).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, mgr.ID, GrantRole, fmt.Sprintf(
			"為 %s 賦予 %s 團隊的 %s 權限",
			uname, mgr.Org.Name, role,
		))
		if err != nil {
			return
		}

		err = LogByUserName(tx, uname, GrantRole, fmt.Sprintf(
			"被 %s 賦予 %s 團隊的 %s 權限", mgr.Name, mgr.Org.Name, role,
		))

		return
	})
}

func OrgRevokeRole(tx *gorm.DB, mgr *User, uname string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var victim User
		err = tx.Where("name = ?", uname).First(&victim).Error

		err = tx.Model(&User{}).
			Where("name = ? ", uname).
			Updates(map[string]any{
				"role":   Member,
				"org_id": nil,
			}).Error
		if err != nil {
			return
		}
		err = tx.Model(&victim).
			Updates(map[string]any{
				"role":   Member,
				"org_id": nil,
			}).Error
		if err != nil {
			return
		}

		// 停用他的排程
		err = tx.Model(&Schedule{}).
			Where("user_id = ?", victim.ID).
			Update("disabled", true).Error
		if err != nil {
			return
		}

		// 取消他的預約
		schedules := tx.Model(&Schedule{}).
			Where("schedules.user_id = ?", victim.ID).
			Select("schedules.id")
		err = tx.Model(&Appointment{}).
			Where("schedule_id IN (?)", schedules).
			Where("status NOT IN = (?,?,?)",
				Cancelled, Completed, Missed).
			Update("status", Cancelled).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, mgr.ID, RevokeRole, fmt.Sprintf(
			"將 %s (%s) 移出 %s 團隊，並取消他的排程和預約",
			victim.Name, victim.Note,
			mgr.Org.Name,
		))
		if err != nil {
			return
		}

		err = LogByUserName(tx, uname, RevokeRole, fmt.Sprintf(
			"被 %s (%s) 移出 %s 團隊，排程與預約也都取消",
			mgr.Name, mgr.Note,
			mgr.Org.Name,
		))
		if err != nil {
			return
		}

		return
	})
}

func UpdateMemberNote(tx *gorm.DB, u, victim *User, note string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(&victim).Update("note", note).Error
		if err != nil {
			return err
		}

		err = LogByUID(tx, u.ID, EditMemberNote, fmt.Sprintf(
			"編輯團隊成員 %s (%s) 的備註",
			victim.Name, victim.Note,
		))
		if err != nil {
			return err
		}

		return LogByUID(tx, victim.ID, EditMemberNote, fmt.Sprintf(
			"被幹部 %s (%s) 編輯備註",
			u.Name, u.Note,
		))
	})
}

func UpdateMemberSecret(tx *gorm.DB, u, victim *User, secretNote string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(&victim).Update("secret_note", secretNote).Error
		if err != nil {
			return err
		}

		err = LogByUID(tx, u.ID, EditMemberSecret, fmt.Sprintf(
			"編輯團隊成員 %s (%s) 的秘密備註",
			victim.Name, victim.Note,
		))
		if err != nil {
			return err
		}

		return
	})
}

func UpdateMemberOTP(tx *gorm.DB, u, victim *User, secret string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(&victim).Update("totp_secret", secret).Error
		if err != nil {
			return err
		}

		err = LogByUID(tx, u.ID, EditMemberOTP, fmt.Sprintf(
			"重設團隊成員 %s (%s) 的雙重驗證",
			victim.Name, victim.Note,
		))
		if err != nil {
			return err
		}

		err = LogByUID(tx, victim.ID, EditMemberOTP, fmt.Sprintf(
			"被幹部 %s (%s) 重設雙重驗證",
			u.Name, u.Note,
		))

		return
	})
}
