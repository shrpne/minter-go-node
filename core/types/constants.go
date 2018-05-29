package types

func GetBaseCoin() CoinSymbol {
	return getBaseCoin(2)
}

func getBaseCoin(chainId int) CoinSymbol {
	var coin CoinSymbol

	switch chainId {
	case 1:
		copy(coin[:], []byte("BIP"))
	case 2:
		copy(coin[:], []byte("MNT"))
	}

	coin[4] = byte(0)

	return coin
}
