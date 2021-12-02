package main

import (
	"bytes"

	"encoding/json"

	"fmt"

	"github.com/jmoiron/sqlx"

	"io"

	"net/http"

	"net/url"

	"strings"

	"utils"
)

func PurchaseTicket(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func(e *error) {
		if e != nil && *e != nil {
			log.Println(*e)
			utils.JSON(w, http.StatusInternalServerError, utils.InternalErrorResponse(*e))
		}
	}(&err)

	var iLoginSessionUserID int
	iLoginSessionUserID, err = getLoginSessionUserID(r)
	if err != nil {
		log.Println(err)
		utils.JSON(w, http.StatusUnauthorized, utils.Response{nil, "login required", utils.ErrCodeUnAuthorized})
		err = nil
		return
	}

	var tx *sqlx.Tx
	tx, err = pg.Beginx()
	if err != nil {
		return
	}
	defer tx.Rollback()
	var CurrentUserVariable UsersModel

	err = tx.Get(&CurrentUserVariable, `
        SELECT avatar_url, id, created_at, updated_at, first_name, last_name, email, password, school, society, phone, title, jcr
        FROM users
        WHERE id = $1 
    `, iLoginSessionUserID)

	if err != nil {
		return
	}

	var PurchaseTicketVariable CustomModelForPurchaseTicketEndpointConstruct

	var dec322 = json.NewDecoder(r.Body)
	if err = dec322.Decode(&PurchaseTicketVariable); err != nil {
		log.Println(err)
		utils.JSON(w, http.StatusBadRequest, utils.InvalidJSONErrorResponse(err))
		err = nil
		return
	}

	var defaultValueCreatedAt384 = "now()"

	var defaultValueUpdatedAt384 = "now()"

	var defaultValueStatus384 = "Pending"

	var defaultValuePaymentIntentId384 = "Pending"

	var defaultValueResponseCode384 = 0

	var defaultValueChargeAmount384 = 0

	var NewChargeLogVariable ChargeLogsModel
	err = tx.Get(&NewChargeLogVariable, `
		INSERT INTO charge_logs
		(user_id, created_at, updated_at, status, payment_intent_id, response_code, charge_amount)
		VALUES($1, $2, $3, $4, $5, $6, $7)RETURNING *
	`, CurrentUserVariable.Id, defaultValueCreatedAt384, defaultValueUpdatedAt384, defaultValueStatus384, defaultValuePaymentIntentId384, defaultValueResponseCode384, defaultValueChargeAmount384)
	if err != nil {
		return
	}

	fmt.Println(PurchaseTicketVariable.Tickets)

	var SumVariable = 0
	for _, SinglePurchaseTicketVariable := range PurchaseTicketVariable.Tickets {
		var FetchedTicketVariable EventTicketTypesModel

		err = tx.Get(&FetchedTicketVariable, `
        SELECT id, created_at, updated_at, event_id, number_of_tickets, ticket_price, name
        FROM event_ticket_types
        WHERE id = $1 
    `, SinglePurchaseTicketVariable.TicketTypeId)

		fmt.Println(FetchedTicketVariable.TicketPrice)
		fmt.Println(SinglePurchaseTicketVariable.Quantity)

		SumVariable = SumVariable + (FetchedTicketVariable.TicketPrice * SinglePurchaseTicketVariable.Quantity)

		if err != nil {
			return
		}

		var defaultValueCreatedAt386 = "now()"

		var defaultValueUpdatedAt386 = "now()"

		_, err = tx.Exec(`
		INSERT INTO ticket_purchases
		(quantity, charge_log_id, price, scheduled_event_id, user_id, ticket_type_id, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`, SinglePurchaseTicketVariable.Quantity, NewChargeLogVariable.Id, FetchedTicketVariable.TicketPrice, PurchaseTicketVariable.ScheduledEventId, CurrentUserVariable.Id, SinglePurchaseTicketVariable.TicketTypeId, defaultValueCreatedAt386, defaultValueUpdatedAt386)
		if err != nil {
			return
		}
	}

	var AuthorizationVariable = "Bearer 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"
	//var AuthorizationVariable = "Bearer sk_test_51Hc7A3CtcrhGM8Re845Kl3uEX5zujRaaCv1RBmJpwkhAaRxxfB7vK2vJhMvhO5toyw3vzWZuduFcvHaZTWanYRz700MiIQyVwj"

	var CardVariable = "card"

	var GbpVariable = "gbp"

	var urlString388 = "https://api.stripe.com/v1/payment_intents"

	var body388 io.Reader

	var requestData388 = make(url.Values, 0)

	requestData388.Add("currency", fmt.Sprint(GbpVariable))

	requestData388.Add("payment_method_types[]", fmt.Sprint(CardVariable))

	//	requestData388.Add("authorization", fmt.Sprint(AuthorizationVariable))

	requestData388.Add("amount", fmt.Sprint(SumVariable))

	body388 = bytes.NewBufferString(requestData388.Encode())

	var req388 *http.Request
	req388, err = http.NewRequest(strings.ToUpper("post"), urlString388, body388)
	if err != nil {
		return
	}

	var reqHeaders388 = make(http.Header, 0)

	reqHeaders388.Add("authorization", fmt.Sprint(AuthorizationVariable))

	req388.Header = reqHeaders388

	var resp388 *http.Response
	resp388, err = http.DefaultClient.Do(req388)
	if err != nil {
		return
	} else if resp388 != nil && resp388.Body != nil {
		defer resp388.Body.Close()
	}

	var StripePaymentIntentVariable StripePaymentIntentConstruct

	dec388 := json.NewDecoder(resp388.Body)
	if err = dec388.Decode(&StripePaymentIntentVariable); err != nil {
		return
	}

	var StripePaymentIntentResponseCodeVariable = resp388.StatusCode

	var OKVariable = 200

	if OKVariable == StripePaymentIntentResponseCodeVariable {
		_, err = tx.Exec(`
        UPDATE charge_logs
        SET response_code = $1, charge_amount = $2, payment_intent_id = $3
        WHERE id = $4
    `, StripePaymentIntentResponseCodeVariable, StripePaymentIntentVariable.Amount, StripePaymentIntentVariable.Id, NewChargeLogVariable.Id)
		if err != nil {
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		return
	}

	utils.JSON(w, http.StatusOK, utils.OKResponse(StripePaymentIntentVariable))
}
