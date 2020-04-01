package api

import "fmt"

const ShopItemNameVar = "itemName"

const BuyItemPath = "/shop/items/buy/%s"
const GetShopItemsPath = "/shop/items/"

var BuyItemsRoute = fmt.Sprintf(BuyItemPath, fmt.Sprintf("{%s}", ShopItemNameVar))
