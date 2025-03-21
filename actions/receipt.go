// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrIncorrectAppointmentStatus = errors.New("incorrect appointment status")

func UsrCreateReceipt(tx *gorm.DB, user *User, aid int64, secret string) (ret *Receipt, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var app Appointment
		err = tx.
			Where("id = ?", aid).
			Where("user_id = ?", user.ID).
			First(&app).Error
		if err != nil {
			return
		}
		if app.Status != Confirmed {
			err = ErrIncorrectAppointmentStatus
		}

		v := &Receipt{
			AppointmentID:   app.ID,
			Secret:          secret,
			CreatedByUserID: user.ID,
			CreatedAt:       time.Now(),
		}
		err = tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "appointment_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"secret", "created_at"}),
			}).
			Create(v).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, user.ID, CreateReceipt, fmt.Sprintf(
			"您建立/修改了簽收條碼，密語為 %s",
			secret,
		))
		if err != nil {
			return
		}

		ret = v
		return
	})

	return
}

func UsrGetReceipt(tx *gorm.DB, user *User, id int64) (ret *Receipt, err error) {
	var x Receipt
	err = tx.Model(&x).
		Joins("JOIN appointments ON appointments.id = receipts.appointment_id").
		Where("receipts.id = ?", id).
		Where("appointments.user_id = ?", user.ID).
		First(&x).Error
	if err != nil {
		return
	}

	ret = &x
	return
}

func VolGetReceipt(tx *gorm.DB, vol *User, id int64) (ret *Receipt, err error) {
	var x Receipt
	err = tx.Model(&x).
		Preload("Manager").
		Preload("Manager.Org").
		Preload("CreatedByUser").
		Preload("CreatedByUser.Org").
		Preload("Appointment").
		Preload("Appointment.User").
		Preload("Appointment.Schedule").
		Preload("Appointment.Schedule.User").
		Preload("Appointment.Schedule.User.Org").
		Where("id = ?", id).
		First(&x).Error
	if err != nil {
		return
	}

	if x.Appointment.Schedule.UserID != vol.ID {
		err = ErrNotMyAppointment
	}

	ret = &x
	return
}

func VolCreateReceipt(tx *gorm.DB, vol *User, aid int64, secret string) (ret *Receipt, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var app Appointment
		err = tx.Model(&app).
			Preload("Schedule").
			Preload("User").
			Where("id = ?", aid).
			First(&app).Error
		if err != nil {
			return
		}
		if app.Status != Confirmed {
			err = ErrIncorrectAppointmentStatus
		}
		if app.Schedule.UserID != vol.ID {
			err = ErrNotMyAppointment
			return
		}

		v := &Receipt{
			AppointmentID:   app.ID,
			Secret:          secret,
			CreatedByUserID: vol.ID,
		}
		err = tx.Create(v).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, vol.ID, CreateReceipt, fmt.Sprintf(
			"你代替民眾 %s 建立/修改了簽收條碼，密語為 %s",
			app.User.Name, secret,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, app.User.ID, CreateReceipt, fmt.Sprintf(
			"志工 %s 代替你建立/修改了簽收條碼，密語為 %s",
			vol.Name, secret,
		))
		if err != nil {
			return
		}

		ret = v
		return
	})

	return
}

func VolSignReceipt(tx *gorm.DB, vol *User, id int64, note string, receives map[string]int) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var rcpt Receipt
		err = tx.Model(&rcpt).
			Preload("Appointment").
			Preload("Appointment.User").
			Preload("Appointment.Schedule").
			Where("id = ?", id).
			First(&rcpt).Error
		if err != nil {
			return
		}
		if rcpt.Appointment.Schedule.UserID != vol.ID {
			err = ErrNotMyAppointment
			return
		}
		if rcpt.Appointment.Status != Confirmed {
			err = ErrIncompatibleStatus
		}

		cnt := 0
		for _, v := range receives {
			cnt += v
		}

		now := time.Now()
		dest := Receipt{
			Note:   note,
			SignAt: &now,
		}
		if receives != nil {
			dest.Receives = receives
		}

		err = tx.Model(&rcpt).Updates(dest).Error
		if err != nil {
			return
		}
		err = tx.Model(rcpt.Appointment).Update("status", Completed).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, vol.ID, SignReceipt, fmt.Sprintf(
			"你簽收了民眾 %s 的 %d 份連署書",
			rcpt.Appointment.User.Name, cnt,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, rcpt.Appointment.User.ID, SignReceipt, fmt.Sprintf(
			"志工 %s 簽收了你的 %d 份連署書",
			vol.Name, cnt,
		))
		return
	})
}
