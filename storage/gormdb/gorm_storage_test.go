package gormdb_test

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/storage/gormdb"
)

func TestGormStorageSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var _ = Describe("Gorm storage test", func() {
	var mock sqlmock.Sqlmock
	var repository storage.Storage

	BeforeEach(func() {
		var err error
		var db *sql.DB
		db, mock, err = sqlmock.New()
		mock.MatchExpectationsInOrder(false)
		Expect(err).ShouldNot(HaveOccurred())
		repository = gormdb.New(&storage.Settings{})

		mock.ExpectQuery(`SELECT CURRENT_DATABASE()`).WillReturnRows(sqlmock.NewRows([]string{"mock"}).FromCSVString("mock"))
		mock.ExpectQuery(`SELECT COUNT(1)*`).WillReturnRows(sqlmock.NewRows([]string{"mock"}).FromCSVString("1"))
		mock.ExpectExec("SELECT pg_advisory_lock*").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT version, dirty FROM "schema_migrations" LIMIT 1`).WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))
		mock.ExpectExec("SELECT pg_advisory_unlock*").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

		repository.Open(func(driver, url string) (*sql.DB, error) {
			return db, nil
		})
		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	Describe("Create", func() {
		It("should insert successfully", func() {
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "merchants"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg())

			repository.Create(&model.Merchant{
				Name: "hello",
			})

			Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		})

		When("model is not registered", func() {
			It("should return an error", func() {
				_, err := repository.Create(&testModel{})
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("Save", func() {
		It("should update successfully", func() {
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta(`UPDATE "merchants"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg())

			repository.Save(&model.Merchant{
				UUID: "someid",
				Name: "hello",
			})

			Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		})

		When("model is not registered", func() {
			It("should return an error", func() {
				err := repository.Save(&testModel{})
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("Get", func() {
		It("should get successfully", func() {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions"`)).
				WithArgs(sqlmock.AnyArg())

			repository.Get(model.TransactionObjectType, "someid")

			Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		})

		When("model is not registered", func() {
			It("should return an error", func() {
				_, err := repository.Get("unknown", "someid")
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("Count", func() {
		It("should count successfully", func() {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "transactions" WHERE (uuid`)).
				WithArgs(sqlmock.AnyArg())

			repository.Count(model.TransactionObjectType, "uuid = ?", "id")

			Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		})

		When("model is not registered", func() {
			It("should return an error", func() {
				_, err := repository.Count("unknown", "uuid = ?", "id")
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

type testModel struct {
}

func (t *testModel) GetType() string {
	return "test"
}

func (t *testModel) Validate() error {
	return nil
}
