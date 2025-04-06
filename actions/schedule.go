// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package actions

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"time"

	"github.com/Ronmi/yugi/config"
	"gorm.io/gorm"
)

func CreateSchedule(tx *gorm.DB, u *User, begin, end time.Time, area string) (ret *Schedule, err error) {
	err = tx.Transaction(func(tx *gorm.DB) (err error) {
		s := &Schedule{
			BeginAt: begin,
			EndAt:   end,
			Area:    area,
			User:    u,
		}

		if err = tx.Create(s).Error; err != nil {
			return
		}

		err = LogByUID(
			tx,
			u.ID,
			NewSchedule,
			"建立了新排程#"+strconv.FormatInt(s.ID, 10),
		)
		if err != nil {
			return
		}

		ret = s
		return nil
	})
	return
}

type TimeSpec struct {
	BeginBefore *time.Time
	BeginAfter  *time.Time
	EndBefore   *time.Time
	EndAfter    *time.Time
	Include     *time.Time
}

const TimeSpecFmt = "2006-01-02T15:04"

func (ts TimeSpec) ApplyTo(q *gorm.DB) *gorm.DB {
	if ts.BeginBefore != nil {
		q = q.Where("begin_at <= ?", *ts.BeginBefore)
	}
	if ts.BeginAfter != nil {
		q = q.Where("begin_at >= ?", *ts.BeginAfter)
	}
	if ts.EndBefore != nil {
		q = q.Where("end_at <= ?", *ts.EndBefore)
	}
	if ts.EndAfter != nil {
		q = q.Where("end_at >= ?", *ts.EndAfter)
	}
	if ts.Include != nil {
		q = q.Where("begin_at <= ?", *ts.Include).Where("end_at >= ?", *ts.Include)
	}
	return q
}

func (ts TimeSpec) SetInclude(t time.Time) TimeSpec {
	ts.Include = &t
	return ts
}

func (ts TimeSpec) SetBeginBefore(t time.Time) TimeSpec {
	ts.BeginBefore = &t
	return ts
}

func (ts TimeSpec) SetBeginAfter(t time.Time) TimeSpec {
	ts.BeginAfter = &t
	return ts
}

func (ts TimeSpec) SetEndBefore(t time.Time) TimeSpec {
	ts.EndBefore = &t
	return ts
}

func (ts TimeSpec) SetEndAfter(t time.Time) TimeSpec {
	ts.EndAfter = &t
	return ts
}

func (ts TimeSpec) NZ() bool {
	return ts.BeginBefore != nil || ts.BeginAfter != nil || ts.EndBefore != nil || ts.EndAfter != nil || ts.Include != nil
}

func (ts TimeSpec) ToQuery() template.URL {
	q := url.Values{}
	if ts.BeginBefore != nil {
		q.Add("begin_before", ts.BeginBefore.Format(TimeSpecFmt))
	}
	if ts.BeginAfter != nil {
		q.Add("begin_after", ts.BeginAfter.Format(TimeSpecFmt))
	}
	if ts.EndBefore != nil {
		q.Add("end_before", ts.EndBefore.Format(TimeSpecFmt))
	}
	if ts.EndAfter != nil {
		q.Add("end_after", ts.EndAfter.Format(TimeSpecFmt))
	}
	if ts.Include != nil {
		q.Add("include", ts.Include.Format(TimeSpecFmt))
	}
	return template.URL(q.Encode())
}

func parseTimeSpec(s string) (ret *time.Time) {
	t, err := time.ParseInLocation(TimeSpecFmt, s, config.TZ)
	if err == nil {
		ret = &t
	}
	return
}

func ParseTimeSpec(bb, ba, eb, ea, inc string) (ret TimeSpec) {
	ret.BeginBefore = parseTimeSpec(bb)
	ret.BeginAfter = parseTimeSpec(ba)
	ret.EndBefore = parseTimeSpec(eb)
	ret.EndAfter = parseTimeSpec(ea)
	ret.Include = parseTimeSpec(inc)
	return
}

func ListOrgSchedule(tx *gorm.DB, org *Org, user string, spec TimeSpec) (ret []*Schedule, err error) {
	q := tx.Model(&Schedule{}).
		Joins("LEFT JOIN users ON users.id = schedules.user_id").
		Preload("User").
		Where("users.org_id = ?", org.ID)
	if user != "" {
		q = q.Where("users.name = ?", user)
	}
	q = spec.ApplyTo(q)

	err = q.Find(&ret).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func ListUserSchedule(tx *gorm.DB, u *User, spec TimeSpec) (ret []*Schedule, err error) {
	q := spec.ApplyTo(tx.Model(&Schedule{}).
		Preload("User").
		Where("user_id = ?", u.ID))

	err = q.Find(&ret).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func ListAvailableSchedule(tx *gorm.DB, now time.Time, spec TimeSpec) (ret []*Schedule, err error) {
	q := spec.ApplyTo(tx.Model(&Schedule{}).
		Where("end_at > ?", now).
		Where("disabled = ?", false))

	err = q.Find(&ret).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func MgrDisableSchedule(tx *gorm.DB, mgr *User, id int64) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		s := &Schedule{}
		if err = tx.Preload("User").First(s, id).Error; err != nil {
			return
		}
		if err = tx.Model(s).Update("disabled", true).Error; err != nil {
			return
		}

		err = LogByUID(tx, mgr.ID, DisableSchedule, fmt.Sprintf(
			"停用了 %s (%s) 的排程#%d",
			s.User.Name, s.User.Note, s.ID,
		))
		if err != nil {
			return
		}

		err = LogByUID(tx, s.User.ID, DisableSchedule, fmt.Sprintf(
			"您的排程#%d 已被幹部 %s (%s) 停用",
			s.ID, mgr.Name, mgr.Note,
		))
		return
	})
}

func VolDisableSchedule(tx *gorm.DB, vol *User, id int64) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		s := &Schedule{}
		err = tx.Where("user_id = ?", vol.ID).
			Where("id = ?", id).
			First(s).Error
		if err != nil {
			return
		}

		if err = tx.Model(s).Update("disabled", true).Error; err != nil {
			return
		}

		err = LogByUID(tx, vol.ID, DisableSchedule, fmt.Sprintf(
			"停用了排程#%d", s.ID,
		))
		return
	})
}
