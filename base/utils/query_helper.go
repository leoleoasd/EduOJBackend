package utils

import (
	"fmt"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/database/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func Paginator(query *gorm.DB, limit, offset int, requestURL *url.URL, output interface{}) (total int, prevUrl, nextUrl *string, err error) {
	if limit == 0 {
		limit = 20 // Default limit
	}
	total64 := int64(0)
	err = query.Count(&total64).Error
	if err != nil {
		err = errors.Wrap(err, "could not query count of objects")
		return
	}
	total = int(total64)
	err = query.Limit(limit).Offset(offset).Find(output).Error
	if err != nil {
		err = errors.Wrap(err, "could not query objects")
		return
	}
	count := reflect.ValueOf(output).Elem().Len()

	if offset-limit >= 0 {
		prevURL := requestURL
		q, err := url.ParseQuery(prevURL.RawQuery)
		if err != nil {
			panic(errors.Wrap(err, "could not parse query for url"))
		}
		q.Set("offset", fmt.Sprint(offset-limit))
		q.Set("limit", fmt.Sprint(limit))
		prevURL.RawQuery = q.Encode()
		temp := prevURL.String()
		prevUrl = &temp
	} else {
		prevUrl = nil
	}
	if offset+count < total {
		nextURL := requestURL
		q, err := url.ParseQuery(nextURL.RawQuery)
		if err != nil {
			panic(errors.Wrap(err, "could not parse query for url"))
		}
		q.Set("offset", fmt.Sprint(offset+limit))
		q.Set("limit", fmt.Sprint(limit))
		nextURL.RawQuery = q.Encode()
		temp := nextURL.String()
		nextUrl = &temp
	} else {
		nextUrl = nil
	}
	return
}

func Sorter(query *gorm.DB, orderBy string, columns ...string) (*gorm.DB, error) {
	if orderBy != "" {
		order := strings.SplitN(orderBy, ".", 2)
		if len(order) != 2 {
			return nil, HttpError{
				Code:    http.StatusBadRequest,
				Message: "INVALID_ORDER",
			}
		}
		if !Contain(order[0], columns) {
			return nil, HttpError{
				Code:    http.StatusBadRequest,
				Message: "INVALID_ORDER",
			}
		}
		if !Contain(order[1], []string{"ASC", "DESC"}) {
			return nil, HttpError{
				Code:    http.StatusBadRequest,
				Message: "INVALID_ORDER",
			}
		}
		query = query.Order(strings.Join(order, " "))
	}
	return query, nil
}

func FindUser(id string) (*models.User, error) {
	user := models.User{}
	err := base.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		err = base.DB.Where("username = ?", id).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			} else {
				return nil, errors.Wrap(err, "could not query user")
			}
		}
	}
	return &user, nil
}

// This function checks if the user has permission to get problems which are not public.
// nil user pointer is regarded as admin(skip the permission judgement).
func FindProblem(id string, user *models.User) (*models.Problem, error) {
	problem := models.Problem{}
	query := base.DB
	err := query.Where("id = ?", id).First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		} else {
			return nil, errors.Wrap(err, "could not query problem")
		}
	}
	if !problem.Public && user != nil && !user.Can("read_problem", problem) {
		return nil, gorm.ErrRecordNotFound
	}
	problem.LoadTestCases()
	return &problem, nil
}

// This function checks if the user has permission to get problems which are not public.
// nil user pointer is regarded as admin(skip the permission judgement).
func FindTestCase(problemId string, testCaseIdStr string, user *models.User) (*models.TestCase, *models.Problem, error) {
	problem, err := FindProblem(problemId, user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	} else if err != nil {
		panic(err)
	}
	testCaseId, err := strconv.ParseUint(testCaseIdStr, 10, 64)
	if err != nil {
		return nil, problem, err
	}
	for _, t := range problem.TestCases {
		if uint64(t.ID) == testCaseId {
			return &t, problem, nil
		}
	}
	return nil, problem, gorm.ErrRecordNotFound
}

func FindSubmission(id uint, includeProblemSet bool) (*models.Submission, error) {
	submission := models.Submission{}
	err := base.DB.Where("id = ?", id).First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		} else {
			return nil, errors.Wrap(err, "could not query submission")
		}
	}
	if !includeProblemSet && submission.ProblemSetId == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &submission, nil
}
