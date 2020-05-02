package api

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/query"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/web"
)

type MerchantService interface {
	Create(*model.Merchant) (model.Object, error)
	Get(string) (*model.Merchant, error)
	List(query []query.Query) ([]model.Object, error)
}

type PagesController struct {
	paymentService  PaymentService
	merchantService MerchantService
}

func NewPagesController(paymentService PaymentService, merchantService MerchantService) web.Controller {
	return &PagesController{
		paymentService:  paymentService,
		merchantService: merchantService,
	}
}

func (c *PagesController) showTransactions(rw http.ResponseWriter, req *web.Request) {
	log.Printf("Requesting transactions page")

	fp := path.Join("templates", "transactions.html")
	funcs := template.FuncMap{
		"ftime": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
	}
	tmpl, err := template.New("transactions.html").Funcs(funcs).ParseFiles(fp)
	if err != nil {
		log.Printf("Could not load view: %s", err)
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not load view",
		})
		return
	}
	ctx := req.Request.Context()
	ctxQuery := query.QueryFromContext(ctx)
	transactions, err := c.paymentService.List(ctxQuery)
	if err != nil {
		web.WriteError(rw, err)
		return
	}

	var merchantID string
	for _, cq := range ctxQuery {
		if cq.Type == model.TransactionObjectType && cq.Key == "merchant_id" {
			merchantID = cq.Value
		}
	}
	merchant, err := c.merchantService.Get(merchantID)
	if err != nil {
		if err == storage.ErrNotFound {
			merchant = nil
		} else {
			web.WriteError(rw, fmt.Errorf("could not get merchant: %s", err))
			return
		}
	}

	transactionPageModel := mapTransactionByParents(transactions)
	user, found := web.UserFromContext(ctx)
	if !found {
		web.WriteError(rw, errors.New("user not found"))
		return
	}

	merchants := make([]model.Object, 0)
	matched, _ := web.HasScopes(user.Scopes, []string{"merchant.read"})
	if matched {
		merchants, err = c.merchantService.List(nil)
		if err != nil {
			web.WriteError(rw, err)
			return
		}
	}

	if err := tmpl.Execute(rw, map[string]interface{}{
		"user":         user,
		"merchant":     merchant,
		"merchants":    merchants,
		"transactions": transactionPageModel,
	}); err != nil {
		log.Printf("Could not execute view: %s", err)
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusInternalServerError,
			Description: "could not execute view",
		})
		return
	}
}

func mapTransactionByParents(transactions []model.Object) map[string][]*model.Transaction {
	children := make(map[string]*model.Transaction)

	for _, t := range transactions {
		tt := t.(*model.Transaction)
		for _, i := range transactions {
			ii := i.(*model.Transaction)
			if ii.DependsOnUUID == tt.UUID {
				children[tt.UUID] = ii
			}
		}
	}

	result := make(map[string][]*model.Transaction)
	for _, t := range transactions {
		tt := t.(*model.Transaction)
		if tt.Type == model.Authorize {
			result[tt.UUID] = make([]*model.Transaction, 0)
			result[tt.UUID] = append(result[tt.UUID], tt)
			currentChild, found := children[tt.UUID]
			for found {
				result[tt.UUID] = append(result[tt.UUID], currentChild)
				currentChild, found = children[currentChild.UUID]
			}
		}
	}

	return result
}

func (c *PagesController) Routes() []web.Route {
	return []web.Route{
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/transactions",
			},
			Handler: c.showTransactions,
			Scopes: func() []string {
				return []string{
					"transaction.read",
				}
			},
		},
		{
			Endpoint: web.Endpoint{
				Method: http.MethodGet,
				Path:   "/",
			},
			Handler: func(rw http.ResponseWriter, req *web.Request) {
				http.Redirect(rw, req.Request, "/templates/resources", http.StatusPermanentRedirect)
			},
		},
	}
}
