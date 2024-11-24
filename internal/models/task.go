package models

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"gorm.io/gorm"
)

type Task struct {
	ID       uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	RoleID   uint64 `gorm:"column:roleId;index:idx_roleId;NOT NULL"`
	TaskId   int32  `gorm:"column:taskId;NOT NULL"`
	TaskType byte   `gorm:"column:taskType;NOT NULL"` // 主线 日常任务
	Progress int32  `gorm:"column:progress"`          // 进度
	Complete int32  `gorm:"column:complete"`          // 完成了多少次
	State    int32  `gorm:"column:state"`             // 可领奖励次数
}

var TaskRepo = new(taskRepo)

type taskRepo struct{}

func (s *taskRepo) Get(ctx context.Context, db *gorm.DB, roleId uint64, taskType int, taskId int32) (*Task, error) {
	task := new(Task)
	err := db.Model(new(Task)).Where("roleId=? and taskType=? and taskId=?", roleId, taskType, taskId).First(task).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return task, errors.NewResponseError(constant.DatabaseError, err)
	}
	return task, err
}

func (s *taskRepo) Save(ctx context.Context, db *gorm.DB, task *Task) error {
	if err := db.Save(task).Error; err != nil {
		return errors.NewResponseError(constant.DatabaseError, err)
	}
	return nil
}

func (s *taskRepo) FindAllByRoleId(ctx context.Context, db *gorm.DB, roleId uint64) ([]*Task, error) {
	tasks := make([]*Task, 0)
	err := db.Where("roleId=?", roleId).Find(&tasks).Error
	return tasks, err
}

func (s *taskRepo) FindAllByRoleIdAndType(ctx context.Context, db *gorm.DB, roleId uint64, taskType byte) ([]*Task, error) {
	tasks := make([]*Task, 0)
	err := db.Where("roleId=? and taskType=?", roleId, taskType).Find(&tasks).Error
	return tasks, err
}
