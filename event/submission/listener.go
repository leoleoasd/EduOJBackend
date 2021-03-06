package submission

import (
	"github.com/EduOJ/backend/database/models"
	"github.com/pkg/errors"
)

func UpdateGrade(r EventArgs) EventRst {
	err := models.UpdateGrade(r)
	return errors.Wrap(err, "could not update grade")
}
