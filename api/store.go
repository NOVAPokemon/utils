package api

import "fmt"

const ShopItemNameVar = "itemName"

const BuyItemPath = "/store/items/buy/%s"
const GetShopItemsPath = "/store/items"

var BuyItemsRoute = fmt.Sprintf(BuyItemPath, fmt.Sprintf("{%s}", ShopItemNameVar))
