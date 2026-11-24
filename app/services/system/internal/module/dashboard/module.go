package dashboard

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewDashboardRP,
	NewDashboardUC,
	NewDashboardService,
)
