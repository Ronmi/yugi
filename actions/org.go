// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import "gorm.io/gorm"

func AddOrg(tx *gorm.DB, org Org, manager string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(&org).Error; err != nil {
			return
		}

		return tx.Model(&User{}).
			Where("name = ?", manager).
			Updates(map[string]any{
				"org_id": org.ID,
				"role":   Manager,
			}).Error
	})
}

func GetOrg(tx *gorm.DB, name string) (ret *Org, cnt int64, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var x Org
		err = tx.Where("name = ?", name).First(&x).Error
		if err != nil {
			return
		}

		err = tx.Model(&User{}).
			Where("org_id = ?", x.ID).
			Count(&cnt).Error
		if err != nil {
			return
		}

		ret = &x
		return
	})

	return
}
