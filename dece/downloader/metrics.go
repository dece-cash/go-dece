// copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/dece-cash/go-dece/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("dece/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("dece/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("dece/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("dece/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("dece/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("dece/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("dece/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("dece/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("dece/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("dece/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("dece/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("dece/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("dece/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("dece/downloader/states/drop", nil)
)
