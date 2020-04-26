package services_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/services"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/storage/storagefakes"
)

var _ = Describe("Payment service", func() {
	var fakeStorage *storagefakes.FakeStorage
	var paymentService *services.PaymentService

	var merchant *model.Merchant
	var authorizeTransaction *model.Transaction
	var chargeTransaction *model.Transaction
	var refundTransaction *model.Transaction
	// var reversalTransaction *model.Transaction

	BeforeEach(func() {
		fakeStorage = &storagefakes.FakeStorage{}
		fakeStorage.TransactionStub = func(fs func(s storage.Storage) error) error {
			return fs(fakeStorage)
		}
		paymentService = services.NewPaymentService(fakeStorage)

		merchant = &model.Merchant{
			UUID:                "1",
			Status:              true,
			Email:               "merchant@email.com",
			Name:                "merchant",
			TotalTransactionSum: 0,
		}

		authorizeTransaction = &model.Transaction{
			UUID:          "some-uuid",
			Status:        model.Approved,
			Type:          model.Authorize,
			Amount:        10,
			CustomerEmail: "user@customer.com",
			CustomerPhone: "000000000",
			MerchantID:    "1",
		}

		chargeTransaction = &model.Transaction{
			UUID:          "some-uuid",
			DependsOnUUID: "parent-uuid",
			Status:        model.Approved,
			Type:          model.Charge,
			Amount:        10,
			CustomerEmail: "user@customer.com",
			CustomerPhone: "000000000",
			MerchantID:    "1",
		}

		refundTransaction = &model.Transaction{
			UUID:          "some-uuid",
			DependsOnUUID: "parent-uuid",
			Status:        model.Approved,
			Type:          model.Refund,
			Amount:        10,
			CustomerEmail: "user@customer.com",
			CustomerPhone: "000000000",
			MerchantID:    "1",
		}

		// reversalTransaction = &model.Transaction{
		// 	UUID:          "some-uuid",
		// 	DependsOnUUID: "parent-uuid",
		// 	Status:        model.Approved,
		// 	Type:          model.Reversal,
		// 	Amount:        10,
		// 	CustomerEmail: "user@customer.com",
		// 	CustomerPhone: "000000000",
		// 	MerchantID:    "1",
		// }
	})

	Describe("Create", func() {
		When("there is a merchant associated with the transaction", func() {
			When("merchant is active", func() {
				BeforeEach(func() {
					fakeStorage.GetReturnsOnCall(0, merchant, nil)
				})

				When("transaction is authorize", func() {
					BeforeEach(func() {
						fakeStorage.CreateReturns(authorizeTransaction, nil)
					})

					It("should be created successfully", func() {
						result, err := paymentService.Create(&model.Transaction{
							Type:          model.Authorize,
							Amount:        10,
							CustomerEmail: "user@customer.com",
							CustomerPhone: "000000000",
							MerchantID:    "1",
						})
						Expect(err).ShouldNot(HaveOccurred())
						transaction := result.(*model.Transaction)
						Expect(transaction.Status).To(Equal(model.Approved))
						Expect(transaction.UUID).To(Not(BeEmpty()))
						Expect(transaction.DependsOnUUID).To(BeEmpty())
					})

					When("create returns an error", func() {
						BeforeEach(func() {
							fakeStorage.CreateReturns(nil, errors.New("error during create"))
						})

						It("should return the error", func() {
							_, err := paymentService.Create(&model.Transaction{
								Type:          model.Authorize,
								Amount:        10,
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).Should(HaveOccurred())
						})
					})
				})

				When("transaction is charge", func() {
					BeforeEach(func() {
						fakeStorage.CreateReturns(chargeTransaction, nil)
					})

					When("parent is not of type authorize", func() {
						BeforeEach(func() {
							fakeStorage.GetReturnsOnCall(1, chargeTransaction, nil)
						})

						It("should fail to create", func() {
							_, err := paymentService.Create(&model.Transaction{
								Type:          model.Charge,
								Amount:        10,
								DependsOnUUID: "parent-uuid",
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).Should(ContainSubstring("parent transaction should be of type authorize"))
						})
					})

					When("parent is authorize, but not approved", func() {
						BeforeEach(func() {
							authorizeTransaction.Status = model.Reversed
							fakeStorage.GetReturnsOnCall(1, authorizeTransaction, nil)
						})

						It("should fail to create", func() {
							_, err := paymentService.Create(&model.Transaction{
								Type:          model.Charge,
								Amount:        10,
								DependsOnUUID: "parent-uuid",
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).Should(ContainSubstring("authorize transaction should be approved, but is reversed"))
						})
					})

					When("parent is authorize", func() {
						BeforeEach(func() {
							fakeStorage.GetReturnsOnCall(1, authorizeTransaction, nil)
							fakeStorage.GetReturnsOnCall(2, merchant, nil)
							fakeStorage.SaveReturns(nil)
						})

						It("should be created successfully", func() {
							result, err := paymentService.Create(&model.Transaction{
								Type:          model.Charge,
								DependsOnUUID: "parent-uuid",
								Amount:        10,
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).ShouldNot(HaveOccurred())
							transaction := result.(*model.Transaction)
							Expect(transaction.Status).To(Equal(model.Approved))
							Expect(transaction.UUID).To(Not(BeEmpty()))
							Expect(transaction.DependsOnUUID).ToNot(BeEmpty())
						})
					})

				})

				When("transaction is refund", func() {
					BeforeEach(func() {
						fakeStorage.CreateReturns(refundTransaction, nil)
					})

					When("parent is not of type charge", func() {
						BeforeEach(func() {
							fakeStorage.GetReturnsOnCall(1, authorizeTransaction, nil)
						})

						It("should fail", func() {
							_, err := paymentService.Create(&model.Transaction{
								Type:          model.Refund,
								DependsOnUUID: "parent-id",
								Amount:        10,
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("parent transaction should be of type charge"))
						})
					})

					When("parent is charge, but no approved", func() {
						BeforeEach(func() {
							chargeTransaction.Status = model.Refunded
							fakeStorage.GetReturnsOnCall(1, chargeTransaction, nil)
						})

						It("should fail", func() {
							_, err := paymentService.Create(&model.Transaction{
								Type:          model.Refund,
								DependsOnUUID: "parent-id",
								Amount:        10,
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("cannot refund charge transaction that is in state refunded"))
						})
					})

					When("parent is charge", func() {
						BeforeEach(func() {
							fakeStorage.GetReturnsOnCall(1, chargeTransaction, nil)
							fakeStorage.GetReturnsOnCall(2, merchant, nil)
							fakeStorage.SaveReturns(nil)
						})

						It("should be successfully created", func() {
							result, err := paymentService.Create(&model.Transaction{
								Type:          model.Refund,
								DependsOnUUID: "parent-id",
								Amount:        10,
								CustomerEmail: "user@customer.com",
								CustomerPhone: "000000000",
								MerchantID:    "1",
							})
							Expect(err).ShouldNot(HaveOccurred())
							transaction := result.(*model.Transaction)
							Expect(transaction.Status).To(Equal(model.Approved))
							Expect(transaction.UUID).To(Not(BeEmpty()))
							Expect(transaction.DependsOnUUID).ToNot(BeEmpty())
						})
					})
				})

				When("transaction is reversal", func() {

				})

				When("the parent transaction is already followed", func() {
					BeforeEach(func() {
						fakeStorage.CountReturns(1, nil)
					})

					It("should fail", func() {
						_, err := paymentService.Create(&model.Transaction{
							Type:          model.Charge,
							DependsOnUUID: "parent-id",
							Amount:        10,
							CustomerEmail: "user@customer.com",
							CustomerPhone: "000000000",
							MerchantID:    "1",
						})
						Expect(err).Should(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("the parent transaction is already followed"))
					})
				})

				When("there is no such parent transaction", func() {
					BeforeEach(func() {
						fakeStorage.GetReturnsOnCall(1, nil, errors.New("no such parent found"))
					})

					It("should fail to create", func() {
						_, err := paymentService.Create(&model.Transaction{
							Type:          model.Charge,
							Amount:        10,
							DependsOnUUID: "no-such-parent",
							CustomerEmail: "user@customer.com",
							CustomerPhone: "000000000",
							MerchantID:    "1",
						})
						Expect(err).Should(HaveOccurred())
						Expect(err.Error()).Should(ContainSubstring("no such parent found"))
					})
				})
			})

			When("merchant is inactive", func() {
				BeforeEach(func() {
					merchant.Status = false
					fakeStorage.GetReturnsOnCall(0, merchant, nil)
					fakeStorage.CreateReturns(authorizeTransaction, nil)
				})

				It("should fail to create transaction", func() {
					_, err := paymentService.Create(&model.Transaction{
						Type:          model.Authorize,
						Amount:        10,
						CustomerEmail: "user@customer.com",
						CustomerPhone: "000000000",
						MerchantID:    "1",
					})
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("merchant with name merchant is not active"))
				})
			})
		})

		When("there is no such merchant", func() {
			BeforeEach(func() {
				fakeStorage.GetReturnsOnCall(0, nil, errors.New("merchant not found"))
			})
			It("should fail to create transaction", func() {
				_, err := paymentService.Create(&model.Transaction{
					Type:          model.Authorize,
					Amount:        10,
					CustomerEmail: "user@customer.com",
					CustomerPhone: "000000000",
					MerchantID:    "1",
				})
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("merchant not found"))
			})
		})
	})
})
