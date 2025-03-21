// This Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"errors"

	"gorm.io/gorm"
)

func LogByUID(tx *gorm.DB, uid int64, action Action, desc string) error {
	return tx.Create(&Log{
		Action:      action,
		UserID:      uid,
		Description: desc,
	}).Error
}

func LogByUserName(tx *gorm.DB, uname string, action Action, desc string) error {
	return tx.Exec("INSERT INTO logs (action, user_id, description, created_at) SELECT ?, id, ?, CURRENT_TIMESTAMP FROM users WHERE name = ?", action, desc, uname).Error
}

func MyLogs(tx *gorm.DB, u *User) (ret []*Log, err error) {
	err = tx.Where("user_id = ?", u.ID).
		Order("created_at DESC").
		Find(&ret).
		Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func OrgLogs(tx *gorm.DB, o *Org) (ret []*Log, err error) {
	q := tx.Model(o).Select("id")
	err = tx.Preload("User").
		Where("user_id IN (?)", q).
		Order("created_at DESC").
		Find(&ret).
		Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
