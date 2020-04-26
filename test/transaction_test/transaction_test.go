package transaction_test

import (
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/test"
)

func TestTransactions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transaction integration tests")
}

var _ = Describe("Transactions", func() {
	var testApp *test.TestApp

	merchant := &model.Merchant{
		UUID:                "1",
		Description:         "Test merchant",
		Name:                "merchant",
		Email:               "merchant@mail.com",
		Status:              true,
		TotalTransactionSum: 0,
	}
	BeforeSuite(func() {
		testApp = test.NewTestApp()
	})

	BeforeEach(func() {
		testApp.Repository.Create(merchant)
	})

	AfterEach(func() {
		testApp.Repository.DeleteAll(model.TransactionObjectType)
		testApp.Repository.DeleteAll(model.MerchantType)
	})

	When("authorize transaction is created", func() {
		var authorizeTransactionID string
		BeforeEach(func() {
			transaction := &model.Transaction{
				Amount:        10,
				CustomerEmail: "email",
				CustomerPhone: "0000000",
				MerchantID:    merchant.UUID,
				Type:          model.Authorize,
			}
			authorizeTransactionID = testApp.Expect.POST("/payment").WithJSON(transaction).Expect().
				Status(http.StatusCreated).JSON().Object().Value("uuid").String().Raw()
		})

		It("should not be able to refund authorize transaction", func() {
			testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
				Amount:        10,
				CustomerEmail: "email",
				CustomerPhone: "0000000",
				MerchantID:    "1",
				Type:          model.Refund,
				DependsOnUUID: authorizeTransactionID,
			}).Expect().Status(http.StatusBadRequest).JSON().Object().Value("description").String().Contains("parent transaction should be of type charge")
		})

		It("should find authorize transaction id db", func() {
			result, err := testApp.Repository.Get(model.TransactionObjectType, authorizeTransactionID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).To(Equal(&model.Transaction{
				UUID:          authorizeTransactionID,
				Status:        model.Approved,
				Amount:        10,
				CustomerEmail: "email",
				CustomerPhone: "0000000",
				MerchantID:    merchant.UUID,
				Type:          model.Authorize,
			}))
		})

		When("charge transaction is created for the authorize one", func() {
			var chargeTransactionID string
			BeforeEach(func() {
				chargeTransactionID = testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
					Amount:        10,
					CustomerEmail: "email",
					CustomerPhone: "0000000",
					MerchantID:    "1",
					Type:          model.Charge,
					DependsOnUUID: authorizeTransactionID,
				}).Expect().Status(http.StatusCreated).JSON().Object().Value("uuid").String().Raw()
			})

			It("should be found in database", func() {
				object, err := testApp.Repository.Get(model.TransactionObjectType, chargeTransactionID)
				Expect(err).ShouldNot(HaveOccurred())
				transaction := object.(*model.Transaction)
				Expect(transaction.Type).To(Equal(model.Charge))
				Expect(transaction.Status).To(Equal(model.Approved))
				Expect(transaction.DependsOnUUID).To(Equal(authorizeTransactionID))
			})

			It("should have given amount to the merchant", func() {
				assertMerchantTotalAmount(testApp.Repository, merchant.UUID, 10)
			})

			It("should not be able to create another charge for the same parent", func() {
				testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
					Amount:        10,
					CustomerEmail: "email",
					CustomerPhone: "0000000",
					MerchantID:    "1",
					Type:          model.Charge,
					DependsOnUUID: authorizeTransactionID,
				}).Expect().Status(http.StatusBadRequest).JSON().Object().
					Value("description").String().Contains("the parent transaction is already followed")

				assertMerchantTotalAmount(testApp.Repository, merchant.UUID, 10)
			})

			It("should not be able to create authorize to depend on charge", func() {
				testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
					Amount:        10,
					CustomerEmail: "email",
					CustomerPhone: "0000000",
					MerchantID:    "1",
					Type:          model.Authorize,
					DependsOnUUID: chargeTransactionID,
				}).Expect().Status(http.StatusBadRequest).JSON().Object().
					Value("description").String().Contains("transaction of type authorize cannot depend on another transaction")
			})

			It("should not be able to create reversal to depend on charge", func() {
				testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
					Amount:        10,
					CustomerEmail: "email",
					CustomerPhone: "0000000",
					MerchantID:    "1",
					Type:          model.Reversal,
					DependsOnUUID: chargeTransactionID,
				}).Expect().Status(http.StatusBadRequest).JSON().Object().
					Value("description").String().Contains("parent transaction should be of type authorize")
			})

			When("charge transaction is refunded", func() {
				var refundTransactionID string
				BeforeEach(func() {
					refundTransactionID = testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
						Amount:        10,
						CustomerEmail: "email",
						CustomerPhone: "0000000",
						MerchantID:    "1",
						Type:          model.Refund,
						DependsOnUUID: chargeTransactionID,
					}).Expect().Status(http.StatusCreated).JSON().Object().Value("uuid").String().Raw()
				})

				It("should be found in database", func() {
					object, err := testApp.Repository.Get(model.TransactionObjectType, refundTransactionID)
					Expect(err).ShouldNot(HaveOccurred())
					transaction := object.(*model.Transaction)
					Expect(transaction.Type).To(Equal(model.Refund))
					Expect(transaction.Status).To(Equal(model.Approved))
					Expect(transaction.DependsOnUUID).To(Equal(chargeTransactionID))
				})

				It("should change the charge to refunded state", func() {
					object, err := testApp.Repository.Get(model.TransactionObjectType, chargeTransactionID)
					Expect(err).ShouldNot(HaveOccurred())
					transaction := object.(*model.Transaction)
					Expect(transaction.Type).To(Equal(model.Charge))
					Expect(transaction.Status).To(Equal(model.Refunded))
					Expect(transaction.DependsOnUUID).To(Equal(authorizeTransactionID))
				})

				It("should have taken amount from the merchant", func() {
					assertMerchantTotalAmount(testApp.Repository, merchant.UUID, 0)
				})

				It("should not be able to refund twice", func() {
					testApp.Expect.POST("/payment").WithJSON(&model.Transaction{
						Amount:        10,
						CustomerEmail: "email",
						CustomerPhone: "0000000",
						MerchantID:    "1",
						Type:          model.Refund,
						DependsOnUUID: chargeTransactionID,
					}).Expect().Status(http.StatusBadRequest).JSON().Object().Value("description").String().Contains("the parent transaction is already followed")
				})
			})
		})
	})
})

func assertMerchantTotalAmount(repository storage.Storage, id string, amount int) {
	object, err := repository.Get(model.MerchantType, id)
	Expect(err).ShouldNot(HaveOccurred())
	merchant := object.(*model.Merchant)
	Expect(merchant.TotalTransactionSum).To(Equal(int64(amount)))
}
