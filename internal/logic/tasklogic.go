package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/models"
	"eggServer/internal/schema"
	"eggServer/pkg/errors"
	"eggServer/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"time"
)

var TaskLogic = new(taskLogic)

type taskLogic struct {
	tables    *cfg.Tables
	taskTypes map[int32][]int32
}

func (s *taskLogic) Init(tables *cfg.Tables) {
	s.tables = tables
	s.taskTypes = make(map[int32][]int32)
	taskConfigList := s.tables.TaskTb.GetDataList()
	for _, v := range taskConfigList {
		s.taskTypes[v.TaskSubId] = append(s.taskTypes[v.TaskSubId], v.TaskId)
	}
}

func (s *taskLogic) Daily(ctx context.Context, db *gorm.DB, role *models.Role) error {
	logger := contextx.FromLogger(ctx)
	taskList, err := models.TaskRepo.FindAllByRoleIdAndType(ctx, db, role.ID, cfg.TaskType_Daily) // 每日任务重置
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("TaskLogic.Daily error:%s", err.Error())
		return err
	}

	if len(taskList) > 0 {
		for _, task := range taskList {
			task.Progress = 0
			task.Complete = 0
			task.State = 0
			err = models.TaskRepo.Save(ctx, db, task)
			if err != nil {
				logger.Errorf("TaskLogic.Daily error:%s", err.Error())
				return err
			}
		}
	}

	return nil
}

// 记录任务进度
func (s *taskLogic) doRecordTaskProgress(ctx context.Context, db *gorm.DB, roleId uint64, taskType int, taskId int32, count int32) error {
	logger := contextx.FromLogger(ctx)
	task, err := models.TaskRepo.Get(ctx, db, roleId, taskType, taskId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("TaskLogic.doRecordTaskProgress error:%s", err.Error())
		return err
	}

	taskConfig := s.tables.TaskTb.Get(taskId)
	if taskConfig.MaxTimes <= task.Complete {
		return nil
	}

	task.RoleID = roleId
	task.TaskId = taskId
	task.Progress = task.Progress + count
	task.TaskType = byte(taskType)
	progress := task.Progress - task.Complete*taskConfig.Need
	if progress >= taskConfig.Need {
		num := progress / taskConfig.Need
		task.State = task.State + num
		task.Complete = task.Complete + num
	}

	if err := models.TaskRepo.Save(ctx, db, task); err != nil {
		logger.Errorf("TaskLogic.doRecordTaskProgress error:%s", err.Error())
		return err
	}
	return nil
}

// RecordTaskProgress 记录任务进度
func (s *taskLogic) RecordTaskProgress(ctx context.Context, db *gorm.DB, roleId uint64, taskSubId int32, count int32) error {
	logger := contextx.FromLogger(ctx)
	taskConfigList := s.taskTypes[taskSubId]
	if taskConfigList == nil {
		logger.Errorf("TaskLogic.RecordTaskProgress error: taskSubId(%d) not in taskTypes", taskSubId)
		return errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	for _, taskId := range taskConfigList {
		taskConfig := s.tables.TaskTb.Get(taskId)
		if taskConfig == nil {
			return errors.NewResponseError(constant.ParametersInvalid, nil)
		}

		if err := s.doRecordTaskProgress(ctx, db, roleId, int(taskConfig.TaskType), taskId, count); err != nil {
			return err
		}
	}
	return nil
}

// GetTaskReward 领取奖励
func (s *taskLogic) GetTaskReward(ctx context.Context, db *gorm.DB, roleId uint64, taskType int, taskId int32) (*schema.RewardData, error) {
	logger := contextx.FromLogger(ctx)
	taskConfig := s.tables.TaskTb.Get(taskId)
	if taskConfig == nil {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	task, err := models.TaskRepo.Get(ctx, db, roleId, taskType, taskId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("TaskLogic.GetTaskReward error:%s", err.Error())
		return nil, err
	}

	if task.State == 0 {
		return nil, errors.NewResponseError(constant.RewardClaimed, err)
	}

	var resp *schema.RewardData
	err = db.Transaction(func(db *gorm.DB) error {
		reward, err := UtilsLogic.AddItem(ctx, db, roleId, taskConfig.Reward.Id, taskConfig.Reward.Num, taskConfig.Reward.Type)
		if err != nil {
			logger.Errorf("TaskLogic.GetTaskReward error:%s", err.Error())
			return err
		}
		resp = reward

		task.RoleID = roleId
		task.TaskId = taskId
		task.TaskType = byte(taskType)
		task.State = task.State - 1
		if err := models.TaskRepo.Save(ctx, db, task); err != nil {
			logger.Errorf("TaskLogic.GetTaskReward error:%s", err.Error())
			return err
		}
		return nil
	})

	return resp, err
}

func (s *taskLogic) GetTaskList(ctx context.Context, db *gorm.DB, roleId uint64) ([]*models.Task, error) {
	logger := contextx.FromLogger(ctx)
	taskList, err := models.TaskRepo.FindAllByRoleId(ctx, db, roleId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("TaskLogic.GetTaskList error:%s", err.Error())
		return taskList, err
	}
	return taskList, nil
}

func (s *taskLogic) RecordTaskProgressWithGoto(ctx context.Context, db *gorm.DB, roleId uint64, userId uint64, userUid string, taskSubId int32, count int32) (*schema.TaskGotoResp, error) {
	taskTypeConfig := s.tables.TaskTypeTb.Get(taskSubId)
	if taskTypeConfig == nil || taskTypeConfig.GotoType == 0 {
		return nil, errors.NewResponseError(constant.ParametersInvalid, nil)
	}

	resp := new(schema.TaskGotoResp)
	resp.Param = taskTypeConfig.Goto

	if taskTypeConfig.GotoType == 1 || taskTypeConfig.GotoType == 3 { // 打开链接记录
		err := s.RecordTaskProgress(ctx, db, roleId, taskSubId, count)
		if err != nil {
			return nil, err
		}
	}

	if taskTypeConfig.GotoType == 2 || taskTypeConfig.GotoType == 4 {
		user, err := models.UserRepo.FindOneByUserId(ctx, db, userId)
		if err != nil {
			return nil, err
		}
		shareFriendsUrl := s.tables.GlobalTb.GetDataList()[0].ShareFriendsUrl
		now := time.Now()
		startParam := fmt.Sprintf("%s%03d%02d%02d00", userUid, user.Platform, now.Month(), now.Day()) // 启动参数
		resp.Param = fmt.Sprintf("%s?startapp=%s", shareFriendsUrl, utils.Encrypt(startParam))
	}

	return resp, nil
}

func (s *taskLogic) Data(ctx context.Context, db *gorm.DB, roleId uint64) (*schema.TaskDataResp, error) {
	taskList, err := s.GetTaskList(ctx, db, roleId)
	if err != nil {
		return nil, err
	}

	resp := new(schema.TaskDataResp)
	resp.Data = make(map[int32]*schema.TaskData)
	for _, task := range taskList {
		data := new(schema.TaskData)
		data.State = task.State
		data.Progress = task.Progress
		data.Complete = task.Complete
		resp.Data[task.TaskId] = data
	}

	return resp, nil
}

func (s *taskLogic) TonAccount(ctx context.Context, db *gorm.DB, roleId uint64, req *schema.TaskTonAccountReq) error {
	if err := s.RecordTaskProgress(ctx, db, roleId, 9, 1); err != nil {
		return err
	}
	return nil
}
