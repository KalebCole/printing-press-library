package parser

import (
	"os"
	"reflect"
	"testing"
)

func TestParseOrderListHiddenIDs(t *testing.T) {
	raw, err := os.ReadFile("testdata/orders_list_hidden_ids.html")
	if err != nil {
		t.Fatal(err)
	}
	page, err := ParseOrderList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Orders) != 3 {
		t.Fatalf("got %d orders, want 3", len(page.Orders))
	}
	wantIDs := []string{"111-1111111-1111111", "D01-2222222-2222222", "333-3333333-3333333"}
	for i, order := range page.Orders {
		if order.OrderID != wantIDs[i] {
			t.Errorf("order %d ID = %q, want %q", i, order.OrderID, wantIDs[i])
		}
	}
	if got := page.Orders[0].DetailURL; got != abs("/your-orders/order-details?orderID=111-1111111-1111111") {
		t.Errorf("detail URL = %q", got)
	}
	for _, order := range page.Orders[:2] {
		if order.PlacedDate != "" || order.Total != 0 || len(order.ItemTitles) != 0 {
			t.Errorf("invented encrypted fields: %+v", order)
		}
	}
	order := page.Orders[2]
	if order.PlacedDate != "2026-11-01" || order.Total != 24.95 || order.Currency != "USD" || !reflect.DeepEqual(order.ItemTitles, []string{"Example notebook"}) || !reflect.DeepEqual(order.ASINs, []string{"B000000001"}) {
		t.Errorf("plaintext fields = %+v", order)
	}
	if !page.HasNext {
		t.Error("expected next page")
	}
}

func TestParseOrderListIDFallback(t *testing.T) {
	for _, tt := range []struct{ name, slot, text, want string }{
		{"legacy", "", "ORDER # 444-4444444-4444444", "444-4444444-4444444"},
		{"slot wins", "amzn1.yourorders.order-card.111-1111111-1111111", "999-9999999-9999999", "111-1111111-1111111"},
		{"invalid slot fallback", "amzn1.yourorders.order-card.invalid", "ORDER # 444-4444444-4444444", "444-4444444-4444444"},
		{"unrelated attribute", "amzn1.other.111-1111111-1111111", "", ""},
		{"trailing junk", "amzn1.yourorders.order-card.111-1111111-1111111-extra", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw := `<div class="order-card" data-csa-c-slot-id="` + tt.slot + `">` + tt.text + `<a href="/gp/buyagain?ue_sid=999-9999999-9999999">Buy it again</a></div>`
			page, err := ParseOrderList([]byte(raw))
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == "" {
				if len(page.Orders) != 0 {
					t.Fatalf("unexpected orders: %+v", page.Orders)
				}
				return
			}
			if len(page.Orders) != 1 || page.Orders[0].OrderID != tt.want {
				t.Fatalf("orders = %+v, want ID %q", page.Orders, tt.want)
			}
		})
	}
}
