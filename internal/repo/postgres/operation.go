package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/pkg/db"
)

type OperationRepo struct {
	client db.Client
}

func NewOperationRepo(client db.Client) *OperationRepo {
	return &OperationRepo{client: client}
}

const (
	operationTypePurchase = "purchase"
	operationTypeTransfer = "transfer"
)

func (r *OperationRepo) ExecPurchaseOperation(ctx context.Context, input entity.PurchaseOperation) error {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Primary()
	}

	insertOperationQuery, args, err := database.QueryBuilder().
		Insert("operations").
		Columns("account_id", "operation_type").
		Values(input.CustomerAccountID, operationTypePurchase).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "BuyItem", QueryRaw: insertOperationQuery}

	var operationID int
	if err := database.QueryRow(ctx, query, args...).Scan(&operationID); err != nil {
		return err
	}

	queryRaw, args, err := database.QueryBuilder().
		Insert("purchase_operations").
		Columns("operation_id", "product_id", "customer_account_id", "quantity", "total_price").
		Values(operationID, input.ItemID, input.CustomerAccountID, input.Quantity, input.TotalPrice).
		ToSql()

	if err != nil {
		return err
	}

	query = db.Query{Name: "BuyItem", QueryRaw: queryRaw}
	if _, err = database.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *OperationRepo) ExecTransferOperation(ctx context.Context, input entity.TransferOperation) error {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Primary()
	}

	insertOperationQuery, args, err := database.QueryBuilder().
		Insert("operations").
		Columns("account_id", "operation_type").
		Values(input.SenderAccountID, operationTypeTransfer).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "ExecTransferOperation", QueryRaw: insertOperationQuery}

	var operationID int
	if err := database.QueryRow(ctx, query, args...).Scan(&operationID); err != nil {
		return err
	}

	queryRaw, args, err := database.QueryBuilder().
		Insert("transfer_operations").
		Columns("operation_id", "sender_account_id", "recipient_account_id", "amount").
		Values(operationID, input.SenderAccountID, input.RecipientAccountID, input.Amount).
		ToSql()

	if err != nil {
		return err
	}

	query = db.Query{Name: "ExecTransferOperation", QueryRaw: queryRaw}
	if _, err = database.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *OperationRepo) GetOutgoingTransfers(ctx context.Context, accountID int) ([]entity.Transfer, error) {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Replica()
	}

	queryRaw, args, err := database.QueryBuilder().
		Select("amount", "users.username as recipient_username").
		From("transfer_operations").
		Join("accounts ON transfer_operations.recipient_account_id = accounts.id").
		Join("users ON accounts.id = users.id").
		Where(sq.Eq{"transfer_operations.sender_account_id": accountID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetSentCoins", QueryRaw: queryRaw}
	rows, err := database.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	sentOp := entity.Transfer{}
	sentOps := []entity.Transfer{}

	for rows.Next() {
		if err = rows.Scan(&sentOp.Amount, &sentOp.RecipientUsername); err != nil {
			return nil, err
		}

		sentOps = append(sentOps, sentOp)
	}

	return sentOps, nil
}

func (r *OperationRepo) GetIncomingTransfers(ctx context.Context, accountID int) ([]entity.Transfer, error) {
	database, ok := ctx.Value(db.DBKey).(db.DB)
	if !ok {
		database = r.client.Replica()
	}

	queryRaw, args, err := database.QueryBuilder().
		Select("amount", "users.username as sender_username").
		From("transfer_operations").
		Join("accounts ON transfer_operations.sender_account_id = accounts.id").
		Join("users ON accounts.id = users.id").
		Where(sq.Eq{"transfer_operations.recipient_account_id": accountID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetReceivedCoins", QueryRaw: queryRaw}
	rows, err := database.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	receivedOp := entity.Transfer{}
	receivedOps := []entity.Transfer{}

	for rows.Next() {
		if err = rows.Scan(&receivedOp.Amount, &receivedOp.SenderUsername); err != nil {
			return nil, err
		}

		receivedOps = append(receivedOps, receivedOp)
	}

	return receivedOps, nil
}
