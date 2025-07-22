package persistence

import (
	"github.com/ODawah/Trading-Insights/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type GormUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepo{db: db}
}

func (r *GormUserRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *GormUserRepo) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *GormUserRepo) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

//func InitPG() {
//	err := godotenv.Load(".env")
//	if err != nil {
//		log.Fatalf("Error loading .env file: %s", err)
//	}
//
//	dsn := fmt.Sprintf(
//		"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
//		os.Getenv("DB_HOST"),
//		os.Getenv("DB_USER"),
//		os.Getenv("DB_PASSWORD"),
//		os.Getenv("DB_NAME"),
//		os.Getenv("DB_PORT"),
//	)
//
//	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//	if err != nil {
//		log.Fatal("Failed to connect to database:", err)
//	}
//	DB = db
//}
