// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrScheduleDisabled    = errors.New("排程已停用")
	ErrTooManyAppointments = errors.New("您已經有一個未完成的預約")
)

func CreateAppointment(tx *gorm.DB, scheduleID int64, user *User, contact ContactMethod) (ret *Appointment, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		// 先檢查 schedule 是否正常
		var s Schedule
		err = tx.Preload("User").First(&s, scheduleID).Error
		if err != nil {
			return
		}
		if s.Disabled {
			return ErrScheduleDisabled
		}

		// 檢查數量，同一個 user 只能有一個未完成的 appointment
		// 未完成 = Pending, Confirmed, NotMatched, Contacting
		var count int64
		err = tx.Model(&Appointment{}).
			Where(
				"user_id = ? AND status IN (?, ?, ?, ?)",
				user.ID,
				Pending, Confirmed, NotMatched, Contacting,
			).
			Count(&count).Error
		if err != nil {
			return
		}

		// 建立 appointment
		err = tx.Create(&Appointment{
			ScheduleID:    scheduleID,
			UserID:        user.ID,
			RegisterAt:    time.Now(),
			ContactMethod: &contact,
			Status:        Pending,
		}).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, s.User.ID, NewAppointment, fmt.Sprintf(
			"民眾 %s 預約了您的排程#%d",
			user.Name, scheduleID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, user.ID, NewAppointment, fmt.Sprintf(
			"您預約了排程#%d",
			scheduleID,
		))
		return
	})

	return
}

func GetUserAppointments(tx *gorm.DB, u *User) (ret []*Appointment, err error) {
	err = tx.
		Preload("Schedule").
		Preload("Schedule.User").
		Preload("Schedule.User.Org").
		Preload("Receipt").
		Where("user_id = ?", u.ID).
		Order("id DESC").
		Find(&ret).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func GetVolAppointments(tx *gorm.DB, vol *User) (ret []*Appointment, err error) {
	tx = tx.
		Preload("Schedule").
		Preload("Schedule.User").
		Preload("Schedule.User.Org").
		Preload("User").
		Preload("User.Org").
		Preload("Receipt").
		Where("schedule_id IN (?)", tx.Model(&Schedule{}).Select("id").Where("user_id = ?", vol.ID))
	if vol.Role == Novice {
		tx = tx.Where(
			"status IN (?, ?, ?, ?)",
			Confirmed, Cancelled, Missed, Completed,
		)
	}
	err = tx.
		Order("id DESC").
		Find(&ret).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func getAppointment(tx *gorm.DB, aid int64) (ret *Appointment, err error) {
	var x Appointment
	err = tx.
		Preload("User").
		Preload("Schedule").
		Preload("Schedule.User").
		Preload("Schedule.User.Org").
		Preload("Receipt").
		Preload("Receipt.CreatedByUser").
		Preload("Receipt.CreatedByUser.Org").
		Preload("Receipt.Manager").
		Preload("Receipt.Manager.Org").
		First(&x, aid).Error
	if err == nil {
		ret = &x
	}

	return
}

func VolGetAppointment(tx *gorm.DB, vol *User, aid int64) (ret *Appointment, err error) {
	ret, err = getAppointment(tx, aid)
	if err != nil {
		return
	}

	if ret.Schedule.UserID != vol.ID {
		return nil, ErrNotMyAppointment
	}

	return
}

func UsrGetAppointment(tx *gorm.DB, user *User, aid int64) (ret *Appointment, err error) {
	ret, err = getAppointment(tx, aid)
	if err != nil {
		return
	}

	if ret.UserID != user.ID {
		return nil, ErrNotMyAppointment
	}

	return
}

func MgrGetAppointment(tx *gorm.DB, mgr *User, aid int64) (ret *Appointment, err error) {
	ret, err = getAppointment(tx, aid)
	if err != nil {
		return
	}

	if *ret.Schedule.User.OrgID != *mgr.OrgID {
		return nil, ErrNotMyAppointment
	}

	return
}

func GetMgrAppointments(tx *gorm.DB, mgr *User) (ret []*Appointment, err error) {
	userQ := tx.Model(&User{}).Select("id").Where("org_id = ?", mgr.OrgID)
	schQ := tx.Model(&Schedule{}).Select("id").Where("user_id IN (?)", userQ)
	err = tx.
		Preload("User").
		Preload("Schedule").
		Preload("Schedule.User").
		Preload("Schedule.User.Org").
		Preload("Receipt").
		Where("schedule_id IN (?)", schQ).
		Order("id DESC").
		Find(&ret).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

var ErrDelNotPending = errors.New("無法取消並非等待中的預約")

func UserDeleteAppointment(tx *gorm.DB, aid int64, user *User) (err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("Schedule").First(&a, aid).Error
		if err != nil {
			return
		}

		if a.Status != Pending {
			return ErrDelNotPending
		}

		err = tx.Model(&a).Update("status", Cancelled).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, user.ID, CancelAppointment, fmt.Sprintf(
			"您取消了預約#%d (行程#%d)",
			aid, a.ScheduleID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.Schedule.UserID, CancelAppointment, fmt.Sprintf(
			"民眾 %s 取消了他的預約#%d (行程#%d)",
			user.Name, aid, a.ScheduleID,
		))

		return
	})
	return
}

var ErrNotMyAppointment = errors.New("這不是您的預約")

func VolDeleteAppointment(tx *gorm.DB, aid int64, vol *User) (err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("User").Preload("Schedule").First(&a, aid).Error
		if err != nil {
			return
		}
		if a.Schedule.UserID != vol.ID {
			return ErrNotMyAppointment
		}

		err = tx.Model(&a).Update("status", Cancelled).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, vol.ID, CancelAppointment, fmt.Sprintf(
			"您取消了預約#%d (行程#%d)",
			aid, a.Schedule.ID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.User.ID, CancelAppointment, fmt.Sprintf(
			"志工 %s 取消了您的預約#%d (行程#%d)",
			vol.Name, aid, a.ScheduleID,
		))

		return
	})
	return
}

func VolEditAppointmentPublicNote(tx *gorm.DB, vol *User, aid int64, note string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("User").
			Preload("Schedule").
			Preload("Schedule.User").
			First(&a, aid).Error
		if err != nil {
			return
		}
		if a.Schedule.UserID != vol.ID && vol.Role != Manager && vol.OrgID != a.Schedule.User.OrgID {
			return ErrNotMyAppointment
		}

		err = tx.Model(&a).Update("user_note", note).Error
		if err != nil {
			return
		}

		if a.Schedule.UserID != vol.ID {
			err = LogByUID(tx, a.Schedule.UserID, EditAppointment, fmt.Sprintf(
				"幹部 %s (%s) 編輯了你的預約#%d (行程#%d) 的備註，原本是 %s",
				vol.Name, vol.Note,
				aid, a.ScheduleID, a.UserNote,
			))
			if err != nil {
				return
			}

			err = LogByUID(tx, vol.ID, EditAppointment, fmt.Sprintf(
				"你編輯了志工 %s (%s) 的預約#%d (行程#%d) 的備註，原本是 %s",
				a.Schedule.User.Name, a.Schedule.User.Note,
				aid, a.ScheduleID, a.UserNote,
			))
		} else {
			err = LogByUID(tx, vol.ID, EditAppointment, fmt.Sprintf(
				"你編輯了預約#%d (行程#%d) 的備註，原本是 %s",
				aid, a.ScheduleID, a.UserNote,
			))
		}
		if err != nil {
			return
		}

		err = LogByUID(tx, a.User.ID, EditAppointment, fmt.Sprintf(
			"志工 %s 編輯了您的預約#%d 的備註 (行程#%d)",
			vol.Name, aid, a.ScheduleID,
		))

		return
	})
}

func VolEditAppointmentSecretNote(tx *gorm.DB, vol *User, aid int64, note string) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("User").
			Preload("Schedule").
			Preload("Schedule.User").
			First(&a, aid).Error
		if err != nil {
			return
		}
		if a.Schedule.UserID != vol.ID && vol.Role != Manager && vol.OrgID != a.Schedule.User.OrgID {
			return ErrNotMyAppointment
		}

		err = tx.Model(&a).Update("volunteer_note", note).Error
		if err != nil {
			return
		}

		if a.Schedule.UserID != vol.ID {
			err = LogByUID(tx, a.Schedule.UserID, EditAppointment, fmt.Sprintf(
				"幹部 %s (%s) 編輯了你的預約#%d (行程#%d) 的秘密備註，原本是 %s",
				vol.Name, vol.Note,
				aid, a.ScheduleID, a.VolunteerNote,
			))
			if err != nil {
				return
			}
			err = LogByUID(tx, vol.ID, EditAppointment, fmt.Sprintf(
				"你編輯了志工 %s (%s) 的預約#%d (行程#%d) 的秘密備註，原本是 %s",
				a.Schedule.User.Name, a.Schedule.User.Note,
				aid, a.ScheduleID, a.VolunteerNote,
			))
		} else {
			err = LogByUID(tx, vol.ID, EditAppointment, fmt.Sprintf(
				"你編輯了預約#%d 的秘密備註 (行程#%d)，原本是 %s",
				aid, a.ScheduleID, a.VolunteerNote,
			))
		}
		if err != nil {
			return
		}

		return
	})
}

var ErrIncompatibleStatus = errors.New("狀態不相容")

func VolSetAppointmentStatus(tx *gorm.DB, vol *User, aid int64, status AppointmentStatus, checks ...AppointmentStatus) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("User").Preload("Schedule").First(&a, aid).Error
		if err != nil {
			return
		}
		if a.Schedule.UserID != vol.ID {
			return ErrNotMyAppointment
		}

		ok := true
		if len(checks) > 0 {
			ok = false
			for _, m := range checks {
				if a.Status == m {
					ok = true
					break
				}
			}
		}
		if !ok {
			return ErrIncompatibleStatus
		}

		err = tx.Model(&a).Update("status", status).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, vol.ID, EditAppointment, fmt.Sprintf(
			"你將預約#%d 的狀態設為 %v (行程#%d)",
			aid, status, a.ScheduleID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.User.ID, EditAppointment, fmt.Sprintf(
			"志工 %s 將您的預約#%d 的狀態設為 %v (行程#%d)",
			vol.Name, aid, status, a.ScheduleID,
		))

		return
	})
}

func MgrSetAppointmentStatus(tx *gorm.DB, mgr *User, aid int64, status AppointmentStatus) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.Preload("User").
			Preload("Schedule").
			Preload("Schedule.User").
			First(&a, aid).Error
		if err != nil {
			return
		}

		if a.Schedule.User.OrgID != mgr.OrgID {
			return ErrNotMyAppointment
		}

		err = tx.Model(&a).Update("status", status).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, mgr.ID, EditAppointment, fmt.Sprintf(
			"你將志工 %s (%s) 的預約#%d (行程#%d) 的狀態設為 %v (原為 %s)",
			a.Schedule.User.Name, a.Schedule.User.Note,
			aid, a.ScheduleID, status, a.Status,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.User.ID, EditAppointment, fmt.Sprintf(
			"志工 %s 將您的預約#%d 的狀態設為 %v (行程#%d)",
			mgr.Name, aid, status, a.ScheduleID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.Schedule.UserID, EditAppointment, fmt.Sprintf(
			"幹部 %s (%s) 將你的預約#%d (行程#%d) 的狀態設為 %v (原為 %s)",
			mgr.Name, mgr.Note,
			aid, a.ScheduleID, status, a.Status,
		))
		return
	})
}

func MgrPairAppointment(tx *gorm.DB, mgr *User, aid int64, sid int64) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		var a Appointment
		err = tx.
			Preload("User").
			Preload("Schedule").
			Preload("Schedule.User").
			First(&a, aid).Error
		if err != nil {
			return
		}
		if *a.Schedule.User.OrgID != *mgr.OrgID {
			return ErrNotMyAppointment
		}

		var s Schedule
		err = tx.
			Preload("User").
			First(&s, sid).Error
		if err != nil {
			return
		}
		if *s.User.OrgID != *mgr.OrgID {
			return ErrNotMyAppointment
		}

		err = tx.Model(&Appointment{}).
			Where("id = ?", aid).
			Update("schedule_id", sid).Error
		if err != nil {
			return
		}

		err = LogByUID(tx, mgr.ID, EditAppointment, fmt.Sprintf(
			"你將 志工 %s (%s) 與民眾 %s 的預約#%d (行程#%d) 重配給行程#%d",
			a.Schedule.User.Name, a.Schedule.User.Note,
			a.User.Name, aid, a.ScheduleID, s.ID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.Schedule.UserID, EditAppointment, fmt.Sprintf(
			"幹部 %s (%s) 將你與民眾 %s 的預約#%d (行程#%d) 重配給行程#%d",
			mgr.Name, mgr.Note,
			a.User.Name, aid, a.ScheduleID, s.ID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, a.User.ID, EditAppointment, fmt.Sprintf(
			"志工 %s 將您的預約#%d (行程#%d) 的行程重配給行程#%d",
			mgr.Name, aid, a.ScheduleID, s.ID,
		))
		return
	})
}
