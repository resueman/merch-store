package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/pkg/db"
)

type OperationRepo struct {
	db db.Client
}

func NewOperationRepo(db db.Client) *OperationRepo {
	return &OperationRepo{db: db}
}

const (
	operationTypePurchase = "purchase"
	operationTypeTransfer = "transfer"
)

func (r *OperationRepo) ExecPurchaseOperation(ctx context.Context, input entity.PurchaseOperation) error {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	insertOperationQuery, args, err := builder.Insert("operations").
		Columns("account_id", "operation_type").
		Values(input.CustomerAccountID, operationTypePurchase).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "BuyItem", QueryRaw: insertOperationQuery}

	var operationID int
	if err := primary.QueryRow(ctx, query, args...).Scan(&operationID); err != nil {
		return err
	}

	queryRaw, args, err := builder.Insert("purchase_operations").
		Columns("operation_id", "product_id", "customer_account_id", "quantity", "total_price").
		Values(operationID, input.ItemID, input.CustomerAccountID, input.Quantity, input.TotalPrice).
		ToSql()

	if err != nil {
		return err
	}

	query = db.Query{Name: "BuyItem", QueryRaw: queryRaw}
	if _, err = primary.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *OperationRepo) ExecTransferOperation(ctx context.Context, input entity.TransferOperation) error {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	insertOperationQuery, args, err := builder.Insert("operations").
		Columns("account_id", "operation_type").
		Values(input.SenderAccountID, operationTypeTransfer).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	query := db.Query{Name: "ExecTransferOperation", QueryRaw: insertOperationQuery}

	var operationID int
	if err := primary.QueryRow(ctx, query, args...).Scan(&operationID); err != nil {
		return err
	}

	queryRaw, args, err := builder.Insert("transfer_operations").
		Columns("operation_id", "sender_account_id", "recipient_account_id", "amount").
		Values(operationID, input.SenderAccountID, input.RecipientAccountID, input.Amount).
		ToSql()

	if err != nil {
		return err
	}

	query = db.Query{Name: "ExecTransferOperation", QueryRaw: queryRaw}
	if _, err = primary.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *OperationRepo) GetOutgoingTransfers(ctx context.Context, accountID int) ([]entity.Transfer, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("amount", "recipient_account_id").
		From("transfer_operations").
		Where(sq.Eq{"sender_account_id": accountID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetSentCoins", QueryRaw: queryRaw}
	rows, err := primary.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	sentOp := entity.Transfer{}
	sentOps := []entity.Transfer{}

	for rows.Next() {
		if err = rows.Scan(&sentOp); err != nil {
			return nil, err
		}

		sentOps = append(sentOps, sentOp)
	}

	return sentOps, nil
}

func (r *OperationRepo) GetIncomingTransfers(ctx context.Context, accountID int) ([]entity.Transfer, error) {
	primary := r.db.Primary()
	builder := primary.QueryBuilder()

	queryRaw, args, err := builder.Select("amount", "sender_account_id").
		From("transfer_operations").
		Where(sq.Eq{"recipient_account_id": accountID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	query := db.Query{Name: "GetReceivedCoins", QueryRaw: queryRaw}
	rows, err := primary.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	receivedOp := entity.Transfer{}
	receivedOps := []entity.Transfer{}

	for rows.Next() {
		if err = rows.Scan(&receivedOp); err != nil {
			return nil, err
		}

		receivedOps = append(receivedOps, receivedOp)
	}

	return receivedOps, nil
}
