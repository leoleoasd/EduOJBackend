package database

import (
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

func GetMigration() *gormigrate.Gormigrate {
	return gormigrate.New(base.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// dummy
		{
			ID: "start",
			Migrate: func(tx *gorm.DB) error {
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
		// create logs table
		{
			ID: "create_logs_table",
			Migrate: func(tx *gorm.DB) error {
				type Log struct {
					ID        uint `gorm:"primaryKey"`
					Level     *int
					Message   string
					Caller    string
					CreatedAt time.Time
				}
				return tx.AutoMigrate(&Log{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("logs")
			},
		},
		// create users table
		{
			ID: "create_users_table",
			Migrate: func(tx *gorm.DB) error {
				type User struct {
					ID       uint   `gorm:"primaryKey" json:"id"`
					Username string `gorm:"uniqueIndex;size:30" json:"username" validate:"required,max=30,min=5"`
					Nickname string `gorm:"index;size:30" json:"nickname"`
					Email    string `gorm:"uniqueIndex;size:320" json:"email"`
					Password string `json:"-"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				return tx.AutoMigrate(&User{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("users")
			},
		},
		// add tokens table
		{
			ID: "create_tokens_table",
			Migrate: func(tx *gorm.DB) error {
				type User struct {
					ID       uint   `gorm:"primaryKey" json:"id"`
					Username string `gorm:"uniqueIndex;size:30" json:"username" validate:"required,max=30,min=5"`
					Nickname string `gorm:"index;size:30" json:"nickname"`
					Email    string `gorm:"uniqueIndex;size:320" json:"email"`
					Password string `json:"-"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				type Token struct {
					ID        uint   `gorm:"primaryKey" json:"id"`
					Token     string `gorm:"unique_index;size:32" json:"token"`
					UserID    uint
					User      User
					CreatedAt time.Time `json:"created_at"`
				}
				return tx.AutoMigrate(&Token{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("tokens")
			},
		},
		// add configs table
		{
			ID: "create_configs_table",
			Migrate: func(tx *gorm.DB) error {
				type Config struct {
					ID        uint    `gorm:"primaryKey"`
					Key       string  `gorm:"size:255"`
					Value     *string `gorm:"default:''"` // 可能是空字符串, 因此得是指针
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.AutoMigrate(&Config{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("configs")
			},
		},
		// add permissions
		{
			ID: "add_permissions",
			Migrate: func(tx *gorm.DB) (err error) {
				type User struct {
					ID       uint   `gorm:"primaryKey" json:"id"`
					Username string `gorm:"uniqueIndex;size:30" json:"username" validate:"required,max=30,min=5"`
					Nickname string `gorm:"index;size:30" json:"nickname"`
					Email    string `gorm:"uniqueIndex;size:320" json:"email"`
					Password string `json:"-"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				type UserHasRole struct {
					ID       uint `gorm:"primaryKey" json:"id"`
					UserID   uint `json:"user_id"`
					User     User
					RoleID   uint `json:"role_id"`
					TargetID uint `json:"target_id"`
				}

				type Role struct {
					ID     uint    `gorm:"primaryKey" json:"id"`
					Name   string  `json:"name" gorm:"size:255"`
					Target *string `json:"target" gorm:"size:255"`
				}

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}
				err = tx.AutoMigrate(&UserHasRole{}, &Role{}, &Permission{})
				if err != nil {
					return
				}
				return
			},
			Rollback: func(tx *gorm.DB) (err error) {
				err = tx.Migrator().DropTable("user_has_roles")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable("permissions")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable("roles")
				if err != nil {
					return
				}
				return
			},
		},
		// add UpdateAt column
		{
			ID: "add_updated_at_column_to_tokens",
			Migrate: func(tx *gorm.DB) error {
				type Token struct {
					UpdatedAt time.Time
				}
				return tx.AutoMigrate(&Token{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Token struct {
					RememberMe bool
				}
				return tx.Migrator().DropColumn(&Token{}, "updated_at")
			},
		},
		// add RememberMe column
		{
			ID: "add_remember_me_column_to_tokens",
			Migrate: func(tx *gorm.DB) error {
				type Token struct {
					RememberMe bool
				}
				return tx.AutoMigrate(&Token{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Token struct {
					RememberMe bool
				}
				return tx.Migrator().DropColumn(&Token{}, "remember_me")
			},
		},
		// add images table
		{
			ID: "create_images_table",
			Migrate: func(tx *gorm.DB) error {
				type Image struct {
					ID        uint      `gorm:"primaryKey" json:"id"`
					Filename  string    `gorm:"filename,size:2048,uniqueIndex"`
					FilePath  string    `gorm:"filepath,size:2048"`
					UserID    uint      `gorm:"index"`
					CreatedAt time.Time `json:"created_at"`
					UpdatedAt time.Time `json:"updated_at"`
				}

				return tx.AutoMigrate(&Image{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("images")
			},
		},
		// add default admin role
		{
			ID: "add_default_admin_role",
			Migrate: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				admin := Role{
					Name:   "admin",
					Target: nil,
				}
				err = tx.Create(&admin).Error
				if err != nil {
					return
				}
				perm := Permission{
					RoleID: admin.ID,
					Name:   "all",
				}
				err = tx.Model(&admin).Association("Permissions").Append(&perm)
				if err != nil {
					return
				}
				return
			},
			Rollback: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				var admin Role
				err = tx.Where("name = ? ", "admin").First(&admin).Error // TODO: and target == nil ?
				if err != nil {
					return
				}
				err = tx.Delete(Permission{}, "role_id = ?", admin.ID).Error
				if err != nil {
					return
				}
				err = tx.Delete(&admin).Error
				if err != nil {
					return
				}
				return
			},
		},
		// add problems
		{
			ID: "add_problems",
			Migrate: func(tx *gorm.DB) (err error) {

				type TestCase struct {
					ID uint `gorm:"primaryKey" json:"id"`

					ProblemID uint `sql:"index" json:"problem_id" gorm:"not null"`
					Score     uint `json:"score" gorm:"default:0;not null"` // 0 for 平均分配

					InputFileName  string `json:"input_file_name" gorm:"size:255;default:'';not null"`
					OutputFileName string `json:"output_file_name" gorm:"size:255;default:'';not null"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}

				type Problem struct {
					ID                 uint   `gorm:"primaryKey" json:"id"`
					Name               string `sql:"index" json:"name" gorm:"size:255;default:'';not null"`
					Description        string `json:"description"`
					AttachmentFileName string `json:"attachment_file_name" gorm:"size:255;default:'';not null"`
					Public             bool   `json:"public" gorm:"default:false;not null"`
					Privacy            bool   `json:"privacy" gorm:"default:false;not null"`

					MemoryLimit        uint64      `json:"memory_limit" gorm:"default:0;not null;type:bigint"`               // Byte
					TimeLimit          uint        `json:"time_limit" gorm:"default:0;not null"`                             // ms
					LanguageAllowed    StringArray `json:"language_allowed" gorm:"size:255;default:'';not null;type:string"` // E.g.    cpp,c,java,python
					CompileEnvironment string      `json:"compile_environment" gorm:"size:2047;default:'';not null"`         // E.g.  O2=false
					CompareScriptName  string      `json:"compare_script_name" gorm:"default:'';not null"`

					TestCases []TestCase `json:"test_cases"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				err = tx.AutoMigrate(&Problem{}, &TestCase{})
				if err != nil {
					return
				}
				return
			},
			Rollback: func(tx *gorm.DB) (err error) {
				err = tx.Migrator().DropTable("test_cases")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable("problems")
				if err != nil {
					return
				}
				return
			},
		},
		{
			// add default problem role
			ID: "add_default_problem_role",
			Migrate: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				problemString := "problem"
				problemCreator := Role{
					Name:   "problem_creator",
					Target: &problemString,
				}
				err = tx.Create(&problemCreator).Error
				if err != nil {
					return
				}
				problemPerm := Permission{
					RoleID: problemCreator.ID,
					Name:   "all",
				}
				err = tx.Model(&problemCreator).Association("Permissions").Append(&problemPerm)
				if err != nil {
					return
				}
				return
			},
			Rollback: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				var problemCreator Role
				err = tx.Where("name = ? and target = ? ", "problem_creator", "problem").First(&problemCreator).Error
				if err != nil {
					return
				}
				err = tx.Delete(Permission{}, "role_id = ?", problemCreator.ID).Error
				if err != nil {
					return
				}
				err = tx.Delete(&problemCreator).Error
				if err != nil {
					return
				}
				return
			},
		},
		{
			// add sample column
			ID: "add_sample_column",
			Migrate: func(tx *gorm.DB) (err error) {
				type TestCase struct {
					Sample bool `json:"sample" gorm:"default:false;not null"`
				}
				return tx.AutoMigrate(&TestCase{})
			},
			Rollback: func(tx *gorm.DB) (err error) {
				type TestCase struct{}
				return tx.Migrator().DropColumn(&TestCase{}, "sample")
			},
		},
		{
			// add submissions
			ID: "add_submissions",
			Migrate: func(tx *gorm.DB) (err error) {
				type User struct {
					ID uint `gorm:"primaryKey" json:"id"`
				}
				type TestCase struct {
					ID uint `gorm:"primaryKey" json:"id"`
				}
				type Problem struct {
					ID uint `gorm:"primaryKey" json:"id"`
				}
				type Run struct {
					ID uint `gorm:"primaryKey" json:"id"`

					UserID       uint      `sql:"index" json:"user_id"`
					User         *User     `json:"user"`
					ProblemID    uint      `sql:"index" json:"problem_id"`
					Problem      *Problem  `json:"problem"`
					ProblemSetId uint      `sql:"index" json:"problem_set_id"`
					TestCaseID   uint      `json:"test_case_id"`
					TestCase     *TestCase `json:"test_case"`
					Sample       bool      `json:"sample" gorm:"not null"`
					SubmissionID uint      `json:"submission_id"`
					Priority     uint8     `json:"priority"`

					Judged             bool   `json:"judged"`
					Status             string `json:"status"`      // AC WA TLE MLE OLE
					MemoryUsed         uint   `json:"memory_used"` // Byte
					TimeUsed           uint   `json:"time_used"`   // ms
					OutputStrippedHash string `json:"output_stripped_hash"`

					CreatedAt time.Time `sql:"index" json:"created_at"`
					UpdatedAt time.Time `json:"-"`
				}
				type Submission struct {
					ID uint `gorm:"primaryKey" json:"id"`

					UserID       uint     `sql:"index" json:"user_id"`
					User         *User    `json:"user"`
					ProblemID    uint     `sql:"index" json:"problem_id"`
					Problem      *Problem `json:"problem"`
					ProblemSetId uint     `sql:"index" json:"problem_set_id"`
					Language     string   `json:"language"`
					FileName     string   `json:"file_name"`
					Priority     uint8    `json:"priority"`

					Judged bool   `json:"judged"`
					Score  uint   `json:"score"`
					Status string `json:"status"`

					Runs []Run `json:"runs"`

					CreatedAt time.Time `sql:"index" json:"created_at"`
					UpdatedAt time.Time `json:"-"`
				}

				return tx.AutoMigrate(&Submission{}, &Run{})
			},
			Rollback: func(tx *gorm.DB) (err error) {
				err = tx.Migrator().DropTable("runs")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable("submissions")
				if err != nil {
					return
				}
				return
			},
		},
		{
			ID: "add_scripts_table",
			Migrate: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.AutoMigrate(&Script{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.Migrator().DropTable(&Script{})
			},
		},
		{
			ID: "add_languages_table",
			Migrate: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				type Language struct {
					Name            string `gorm:"primaryKey"`
					BuildScriptName string
					BuildScript     *Script `gorm:"foreignKey:BuildScriptName"`
					RunScriptName   string
					RunScript       *Script `gorm:"foreignKey:RunScriptName"`
					CreatedAt       time.Time
					UpdatedAt       time.Time
				}
				return tx.AutoMigrate(&Language{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				type Language struct {
					Name            string `gorm:"primaryKey"`
					BuildScriptName string
					BuildScript     *Script `gorm:"foreignKey:BuildScriptName"`
					RunScriptName   string
					RunScript       *Script `gorm:"foreignKey:RunScriptName"`
					CreatedAt       time.Time
					UpdatedAt       time.Time
				}
				return tx.Migrator().DropTable(&Language{})
			},
		},
		{
			ID: "add_fk_submission_language_table",
			Migrate: func(tx *gorm.DB) error {
				type Language struct {
					Name string `gorm:"primaryKey"`
				}

				type Submission struct {
					ID           uint      `gorm:"primaryKey" json:"id"`
					LanguageName string    `json:"language_name"`
					Language     *Language `gorm:"foreignKey:LanguageName"`
				}
				err := tx.Migrator().RenameColumn(&Submission{}, "language", "language_name")
				if err != nil {
					return err
				}
				return tx.AutoMigrate(&Submission{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Language struct {
					Name string `gorm:"primaryKey"`
				}
				type Submission struct {
					ID           uint      `gorm:"primaryKey" json:"id"`
					LanguageName string    `json:"language_name"`
					Language     *Language `gorm:"foreignKey:LanguageName"`
				}
				err := tx.Migrator().RenameColumn(&Submission{}, "language_name", "language")
				if err != nil {
					return err
				}
				return tx.Migrator().DropConstraint(&Submission{}, "fk_submissions_language")
			},
		},
		{
			ID: "add_extension_allowed_Field",
			Migrate: func(tx *gorm.DB) error {
				type Language struct {
					Name             string      `gorm:"primaryKey"`
					ExtensionAllowed StringArray `gorm:"type:string"`
				}
				return tx.AutoMigrate(&Language{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Language struct {
					Name             string      `gorm:"primaryKey"`
					ExtensionAllowed StringArray `gorm:"type:string"`
				}
				return tx.Migrator().DropColumn(&Language{}, "ExtensionAllowed")
			},
		},
		{
			ID: "add_filename_field_to_scripts_table",
			Migrate: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					Filename  string
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.AutoMigrate(&Script{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Script struct {
					Name      string `gorm:"primaryKey"`
					Filename  string
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.Migrator().DropColumn(&Script{}, "Filename")
			},
		},
		{
			ID: "add_judger_field_to_runs_table",
			Migrate: func(tx *gorm.DB) error {

				type Run struct {
					ID            uint `gorm:"primaryKey" json:"id"`
					JudgerName    string
					JudgerMessage string
				}
				return tx.AutoMigrate(&Run{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Run struct {
					ID         uint `gorm:"primaryKey" json:"id"`
					JudgerName string
				}
				if err := tx.Migrator().DropColumn(&Run{}, "JudgerName"); err != nil {
					return err
				}
				return tx.Migrator().DropColumn(&Run{}, "JudgerMessage")
			},
		},
		{
			ID: "add_fk_script_problem",
			Migrate: func(tx *gorm.DB) error {
				type TestCase struct {
					ID uint `gorm:"primaryKey" json:"id"`

					ProblemID uint `sql:"index" json:"problem_id" gorm:"not null"`
					Score     uint `json:"score" gorm:"default:0;not null"` // 0 for 平均分配
					Sample    bool `json:"sample" gorm:"default:false;not null"`

					InputFileName  string `json:"input_file_name" gorm:"size:255;default:'';not null"`
					OutputFileName string `json:"output_file_name" gorm:"size:255;default:'';not null"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"updated_at"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type Script struct {
					Name      string `gorm:"primaryKey"`
					Filename  string
					CreatedAt time.Time
					UpdatedAt time.Time
				}

				type Problem struct {
					ID                 uint   `gorm:"primaryKey" json:"id"`
					Name               string `sql:"index" json:"name" gorm:"size:255;default:'';not null"`
					Description        string `json:"description"`
					AttachmentFileName string `json:"attachment_file_name" gorm:"size:255;default:'';not null"`
					Public             bool   `json:"public" gorm:"default:false;not null"`
					Privacy            bool   `json:"privacy" gorm:"default:false;not null"`

					MemoryLimit        uint64      `json:"memory_limit" gorm:"default:0;not null;type:bigint"`               // Byte
					TimeLimit          uint        `json:"time_limit" gorm:"default:0;not null"`                             // ms
					LanguageAllowed    StringArray `json:"language_allowed" gorm:"size:255;default:'';not null;type:string"` // E.g.    cpp,c,java,python
					CompileEnvironment string      `json:"compile_environment" gorm:"size:2047;default:'';not null"`         // E.g.  O2=false
					CompareScriptName  string      `json:"compare_script_name" gorm:"default:0;not null"`
					CompareScript      Script      `json:"compare_script"`

					TestCases []TestCase `json:"test_cases"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				return tx.AutoMigrate(&Problem{})
			},
			Rollback: func(tx *gorm.DB) error {
				type TestCase struct {
					ID uint `gorm:"primaryKey" json:"id"`

					ProblemID uint `sql:"index" json:"problem_id" gorm:"not null"`
					Score     uint `json:"score" gorm:"default:0;not null"` // 0 for 平均分配
					Sample    bool `json:"sample" gorm:"default:false;not null"`

					InputFileName  string `json:"input_file_name" gorm:"size:255;default:'';not null"`
					OutputFileName string `json:"output_file_name" gorm:"size:255;default:'';not null"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"updated_at"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type Script struct {
					Name      string `gorm:"primaryKey"`
					Filename  string
					CreatedAt time.Time
					UpdatedAt time.Time
				}

				type Problem struct {
					ID                 uint   `gorm:"primaryKey" json:"id"`
					Name               string `sql:"index" json:"name" gorm:"size:255;default:'';not null"`
					Description        string `json:"description"`
					AttachmentFileName string `json:"attachment_file_name" gorm:"size:255;default:'';not null"`
					Public             bool   `json:"public" gorm:"default:false;not null"`
					Privacy            bool   `json:"privacy" gorm:"default:false;not null"`

					MemoryLimit        uint64      `json:"memory_limit" gorm:"default:0;not null;type:bigint"`               // Byte
					TimeLimit          uint        `json:"time_limit" gorm:"default:0;not null"`                             // ms
					LanguageAllowed    StringArray `json:"language_allowed" gorm:"size:255;default:'';not null;type:string"` // E.g.    cpp,c,java,python
					CompileEnvironment string      `json:"compile_environment" gorm:"size:2047;default:'';not null"`         // E.g.  O2=false
					CompareScriptName  string      `json:"compare_script_name" gorm:"default:0;not null"`
					CompareScript      Script      `json:"compare_script"`

					TestCases []TestCase `json:"test_cases"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				return tx.Migrator().DropConstraint(&Problem{}, "fk_submissions_language")
			},
		},
		{
			ID: "rename_problem_compile_envirionment",
			Migrate: func(tx *gorm.DB) error {

				type Problem struct {
					ID                 uint   `gorm:"primaryKey" json:"id"`
					Name               string `sql:"index" json:"name" gorm:"size:255;default:'';not null"`
					Description        string `json:"description"`
					AttachmentFileName string `json:"attachment_file_name" gorm:"size:255;default:'';not null"`
					Public             bool   `json:"public" gorm:"default:false;not null"`
					Privacy            bool   `json:"privacy" gorm:"default:false;not null"`

					MemoryLimit        uint64      `json:"memory_limit" gorm:"default:0;not null;type:bigint"`               // Byte
					TimeLimit          uint        `json:"time_limit" gorm:"default:0;not null"`                             // ms
					LanguageAllowed    StringArray `json:"language_allowed" gorm:"size:255;default:'';not null;type:string"` // E.g.    cpp,c,java,python
					CompileEnvironment string      `json:"compile_environment" gorm:"size:2047;default:'';not null"`         // E.g.  O2=false
					CompareScriptName  string      `json:"compare_script_name" gorm:"default:0;not null"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				return tx.Migrator().RenameColumn(&Problem{}, "compile_environment", "build_arg")
			},
			Rollback: func(tx *gorm.DB) error {
				type Problem struct {
					ID                 uint   `gorm:"primaryKey" json:"id"`
					Name               string `sql:"index" json:"name" gorm:"size:255;default:'';not null"`
					Description        string `json:"description"`
					AttachmentFileName string `json:"attachment_file_name" gorm:"size:255;default:'';not null"`
					Public             bool   `json:"public" gorm:"default:false;not null"`
					Privacy            bool   `json:"privacy" gorm:"default:false;not null"`

					MemoryLimit        uint64      `json:"memory_limit" gorm:"default:0;not null;type:bigint"`               // Byte
					TimeLimit          uint        `json:"time_limit" gorm:"default:0;not null"`                             // ms
					LanguageAllowed    StringArray `json:"language_allowed" gorm:"size:255;default:'';not null;type:string"` // E.g.    cpp,c,java,python
					CompileEnvironment string      `json:"compile_environment" gorm:"size:2047;default:'';not null"`         // E.g.  O2=false
					CompareScriptName  string      `json:"compare_script_name" gorm:"default:0;not null"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				return tx.Migrator().RenameColumn(&Problem{}, "build_arg", "compile_environment")
			},
		},
		{
			// add classes
			ID: "add_classes_table",
			Migrate: func(tx *gorm.DB) error {
				type User struct {
					ID       uint   `gorm:"primaryKey" json:"id"`
					Username string `gorm:"uniqueIndex;size:30" json:"username" validate:"required,max=30,min=5"`
					Nickname string `gorm:"index;size:30" json:"nickname"`
					Email    string `gorm:"uniqueIndex;size:320" json:"email"`
					Password string `json:"-"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				type Class struct {
					ID          uint   `gorm:"primaryKey" json:"id"`
					Name        string `json:"name" gorm:"size:255;default:'';not null"`
					CourseName  string `json:"course_name" gorm:"size:255;default:'';not null"`
					Description string `json:"description"`
					InviteCode  string `json:"invite_code" gorm:"size:255;default:'';not null"`
					Managers    []User `json:"managers" gorm:"many2many:user_manage_classes"`
					Students    []User `json:"students" gorm:"many2many:user_in_classes"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				return tx.AutoMigrate(&Class{})
			},
			Rollback: func(tx *gorm.DB) (err error) {
				type User struct {
					ID       uint   `gorm:"primaryKey" json:"id"`
					Username string `gorm:"uniqueIndex;size:30" json:"username" validate:"required,max=30,min=5"`
					Nickname string `gorm:"index;size:30" json:"nickname"`
					Email    string `gorm:"uniqueIndex;size:320" json:"email"`
					Password string `json:"-"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				type Class struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255;default:'';not null"`
					CourseName  string  `json:"course_name" gorm:"size:255;default:'';not null"`
					Description string  `json:"description"`
					InviteCode  string  `json:"invite_code" gorm:"size:255;default:'';not null;uniqueIndex"`
					Managers    []*User `json:"managers" gorm:"many2many:user_manage_classes"`
					Students    []*User `json:"students" gorm:"many2many:user_in_classes"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
				}
				err = tx.Migrator().DropTable("user_manage_classes")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable("user_in_classes")
				if err != nil {
					return
				}
				return tx.Migrator().DropTable(&Class{})
			},
		},
		{
			ID: "add_WebauthnCredential_table",
			Migrate: func(tx *gorm.DB) error {
				type WebauthnCredential struct {
					ID        uint `gorm:"primaryKey" json:"id"`
					UserID    uint
					Content   string
					CreatedAt time.Time `json:"created_at"`
				}
				return tx.AutoMigrate(WebauthnCredential{})
			},
			Rollback: func(tx *gorm.DB) error {
				type WebauthnCredential struct {
					ID        uint `gorm:"primaryKey" json:"id"`
					UserID    uint
					Content   string
					CreatedAt time.Time `json:"created_at"`
				}
				return tx.Migrator().DropTable(WebauthnCredential{})
			},
		},
		{
			// add default class role
			ID: "add_default_class_role",
			Migrate: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				classString := "class"
				classCreator := Role{
					Name:   "class_creator",
					Target: &classString,
				}
				err = tx.Create(&classCreator).Error
				if err != nil {
					return
				}
				classPerm := Permission{
					RoleID: classCreator.ID,
					Name:   "all",
				}
				err = tx.Model(&classCreator).Association("Permissions").Append(&classPerm)
				if err != nil {
					return
				}
				return
			},
			Rollback: func(tx *gorm.DB) (err error) {

				type Permission struct {
					ID     uint   `gorm:"primaryKey" json:"id"`
					RoleID uint   `json:"role_id"`
					Name   string `json:"name" gorm:"size:255"`
				}

				type Role struct {
					ID          uint    `gorm:"primaryKey" json:"id"`
					Name        string  `json:"name" gorm:"size:255"`
					Target      *string `json:"target" gorm:"size:255"`
					Permissions []Permission
				}

				var classCreator Role
				err = tx.Where("name = ? and target = ? ", "class_creator", "class").First(&classCreator).Error
				if err != nil {
					return
				}
				err = tx.Delete(Permission{}, "role_id = ?", classCreator.ID).Error
				if err != nil {
					return
				}
				err = tx.Delete(&classCreator).Error
				if err != nil {
					return
				}
				return
			},
		},
		{
			ID: "add_deleted_at_run_submission",
			Migrate: func(tx *gorm.DB) error {
				type Submission struct {
					ID        uint           `gorm:"primaryKey" json:"id"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type Run struct {
					ID        uint           `gorm:"primaryKey" json:"id"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				return tx.AutoMigrate(&Submission{}, &Run{})
			},
			Rollback: func(tx *gorm.DB) error {
				type Submission struct {
					ID        uint           `gorm:"primaryKey" json:"id"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type Run struct {
					ID        uint           `gorm:"primaryKey" json:"id"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				if err := tx.Migrator().DropColumn(&Submission{}, "deleted_at"); err != nil {
					return err
				}
				return tx.Migrator().DropColumn(&Run{}, "deleted_at")
			},
		},
		{
			// add problem sets
			ID: "add_problem_set",
			Migrate: func(tx *gorm.DB) error {
				type Problem struct {
					ID uint `gorm:"primaryKey" json:"id"`
				}
				type Grade struct {
					ID uint `gorm:"primaryKey" json:"id"`

					UserID       uint `json:"user_id"`
					ProblemSetID uint `json:"problem_set_id"`

					Detail datatypes.JSON `json:"detail"`
					Total  uint           `json:"total"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type ProblemSet struct {
					ID uint `gorm:"primaryKey" json:"id"`

					ClassID     uint   `sql:"index" json:"class_id" gorm:"not null"`
					Name        string `json:"name" gorm:"not null;size:255"`
					Description string `json:"description"`

					Problems []Problem `json:"problems" gorm:"many2many:problems_in_problem_sets"`
					Grades   []Grade   `json:"grades"`

					StartTime time.Time `json:"start_time"`
					EndTime   time.Time `json:"end_time"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type User struct {
					ID     uint    `gorm:"primaryKey" json:"id"`
					Grades []Grade `json:"grades"`
				}
				type Class struct {
					ID          uint         `gorm:"primaryKey" json:"id"`
					ProblemSets []ProblemSet `json:"problem_sets"`
				}
				return tx.AutoMigrate(&ProblemSet{}, &Grade{}, &User{}, &Class{})
			},
			Rollback: func(tx *gorm.DB) (err error) {
				type Problem struct {
					ID uint `gorm:"primaryKey" json:"id"`
				}
				type Grade struct {
					ID uint `gorm:"primaryKey" json:"id"`

					UserID       uint `json:"user_id"`
					ProblemSetID uint `json:"problem_set_id"`

					Detail datatypes.JSON `json:"detail"`
					Total  uint           `json:"total"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type ProblemSet struct {
					ID uint `gorm:"primaryKey" json:"id"`

					ClassID     uint   `sql:"index" json:"class_id" gorm:"not null"`
					Name        string `json:"name" gorm:"not null;size:255"`
					Description string `json:"description"`

					Problems []Problem `json:"problems" gorm:"many2many:problems_in_problem_sets"`
					Grades   []Grade   `json:"grades"`

					StartTime time.Time `json:"start_time"`
					EndTime   time.Time `json:"end_time"`

					CreatedAt time.Time      `json:"created_at"`
					UpdatedAt time.Time      `json:"-"`
					DeletedAt gorm.DeletedAt `json:"deleted_at"`
				}
				type User struct {
					ID     uint    `gorm:"primaryKey" json:"id"`
					Grades []Grade `json:"grades"`
				}
				type Class struct {
					ID          uint         `gorm:"primaryKey" json:"id"`
					ProblemSets []ProblemSet `json:"problem_sets"`
				}
				err = tx.Migrator().DropTable("problems_in_problem_sets")
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable(&Grade{})
				if err != nil {
					return
				}
				err = tx.Migrator().DropTable(&ProblemSet{})
				if err != nil {
					return
				}
				return
			},
		},
	})
}

func Migrate() {
	m := GetMigration()
	if err := m.Migrate(); err != nil {
		fmt.Printf("Could not migrate: %v", err)
		panic(err)
	}
}
