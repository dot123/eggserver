package models

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		new(User),
		new(Role),
		new(Invite),
		new(Guide),
		new(Item),
		new(Pet),
		new(Egg),
		new(Task),
		new(Shop),
		new(Order),
		new(BattleResult),
		new(Sign),
		new(PassPort),
	)
	// 设置自增起始值
	err = db.Exec("ALTER TABLE g_role AUTO_INCREMENT = 10001;").Error
	if err != nil {
		panic("failed to set AUTO_INCREMENT")
	}
	// 设置自增起始值
	err = db.Exec("ALTER TABLE g_user AUTO_INCREMENT = 10001;").Error
	if err != nil {
		panic("failed to set AUTO_INCREMENT")
	}
	return err
}

// GetPages 分页返回数据
func GetPages(db *gorm.DB, out interface{}, pageNum, pageSize int) (int64, error) {
	var count int64

	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	} else if count == 0 {
		return count, nil
	}

	return count, db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(out).Error
}
