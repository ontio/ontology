/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package errors

import "errors"

var (
	ERR_ASSET_NAME_INVALID        = errors.New("asset name invalid, too long")
	ERR_ASSET_PRECISION_INVALID   = errors.New("asset precision invalid")
	ERR_ASSET_AMOUNT_INVALID      = errors.New("asset amount invalid")
	ERR_ASSET_CHECK_OWNER_INVALID = errors.New("asset owner invalid")
)
