package generate_1

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dece-cash/go-dece/czero/superzk"
	"github.com/dece-cash/go-dece/zero/txtool"
)

func TestSign(t *testing.T) {
	superzk.ZeroInit_NoCircuit()

	str := "{\"Cmds\": {   \"BuyShare\": null,   \"ClosePool\": null,   \"Contract\": null,   \"PkgClose\": null,   \"PkgCreate\": null,   \"PkgTransfer\": null,   \"RegistPool\": null  },  \"Fee\": {   \"Currency\": \"0x000000000000000000000000000000000000000000000000000000005345524f\",   \"Value\": \"25000000000000\"  },  \"From\": {   \"PKr\": \"0xc200d439f05f4cd23e367abc41c505e0f79990a1eee5613cd2033dee36077511b34214e3e0786e14536d22f0b2d7e0068bec6284b626e46f59f974b21d371020948ab05d5c21304e4c4963375c24736670b63a025c709f0b6b97f04a6c1b8b2c\",   \"SKr\": \"0x38da86e05a64c7020db02312130e832e299d96830e7455f6b62da898fc044901ea6b778dfe5206460b35fa4ac75a65fa651c3b8cb5ffcc65d38e5c0a3d0bc9010000000000000000000000000000000000000000000000000000000000000000\"  },  \"Gas\": 25000,  \"GasPrice\": 1000000000,  \"Ins\": [{   \"A\": null,   \"Ar\": null,   \"ArOld\": null,   \"CC\": null,   \"Out\": {    \"Root\": \"0x5122d953535832ab6f2549fb07c1a48301e417acf08a471d5b873552d0304916\",    \"State\": {     \"Num\": 29,     \"OS\": {      \"Index\": 29,      \"OutCM\": null,      \"Out_C\": null,      \"Out_O\": null,      \"Out_P\": null,      \"Out_Z\": {       \"AssetCM\": \"0x7d02ba2fd5b686e76603cc2c14b8c400013611484dae4f534a2753a89e768208\",       \"EInfo\": \"0x172c7962c348449ce8fcea9fa4279eaed4eeb7717ecd669d08c6ba527ac74754d3e64984b54cdbea6c132c08b65494de41e26703b9663de96842f481b46f8f656c8f064f99bf8d1b0c1e813a988f23075d6e591c055f9b814a48dbfc4bc262ea8c9a2cfa5447901c33d9d75be717f0c729f9825ffbc0924e06982279ebfb02969a586894c68a088c673f117019c6c551ca54bdb174c4da2a2c9c86d1dfc26c4d6c5d983343db87cede81b92ad181c6ee58df4658cd65b4da873cb9413eb812c34c38666de024fcfc781f3e9989563e8a822ddf72b1d250b99c68adf54c105acd\",       \"OutCM\": \"0xaa39041683a7b143c22d678daea4aab713e34e928799587464832acf55253590\",       \"PKr\": \"0x3e9a783dc651deca5a801c0a5e490c4fbb82fb0ea0d034642578443f6f879c955537c687923f76efc82abf3dca117e31bcb5df2adf29aa1236a36855987f66a28d00fc36daca0488d0d8b4c041ed954e286ea94198f32a1e76012440f3e18c82\",       \"Proof\": \"0x0214ce94b025a44e68333c8a05e5a63a5141c3042d35b96bef9980df81180adf000a62802b1f7d9435b1a6bca401a9258335805a1c75b488f7e8e8a12ccb29ecd87c98c29974fe2ac337c250f9b5f1407d540040a7b569170a1943779b5de22776040249b0b6bfad004385bec0573d6f77ede1c2ae28bf739caca7df05ff45ec3a2c20\",       \"RPK\": \"0xc58c5fd31db5f4f2b65632a8e916f59bbf354e35b70abef79a35b2ac6fad43c1\"      },      \"RootCM\": \"0x2e78f95db341d04d5e4b6f9201eff7453d8860d84a7c804bca42b5d1ef91f112\"     },     \"TxHash\": \"0xfa98a9c4a2dae32d4f9b90c7f263352a253018fae4330a559077243b6328d77e\"    }   },   \"SKr\": \"0x38da86e05a64c7020db02312130e832e299d96830e7455f6b62da898fc044901ea6b778dfe5206460b35fa4ac75a65fa651c3b8cb5ffcc65d38e5c0a3d0bc9010000000000000000000000000000000000000000000000000000000000000000\",   \"Vskr\": null,   \"Witness\": {    \"Anchor\": \"0x868bb7990cc2423d46eed1eac70a85acbf98c32f48194e3602bb118607509e10\",    \"Paths\": [\"0x4816b0f4f85ac8c25279424b08d8662d831a641e4a31a03bfd0aa083678ab185\", \"0x40163431945acd58343ea4e5a7ea4d1a256c767ebb3ab669975e53da93f40d8c\", \"0xabb1f1d8cc981b4e6c4d75f6343c7d20923fd1531478e00e1912be084bfd6218\", \"0xad15866081d2a25f42e6ba874b2167184dd83013ceb21a5a8edffd09f8771a96\", \"0x2676d074fd027c9ea610408ca619771fe38c653ff024f2f3aca637e5c95b020e\", \"0x75720bad848dc8d1909c2a7201564f6ea10434a565b364a3c4c9328b78d8bc05\", \"0x6eb9d1bbaad7ae8e1ad41b2be597da14321efdbc9c7f4358e6525789d4318c8b\", \"0xa0f4e174f7fe46368a2c4ed56659c672e0187d0bca2e7619660b9e38564a958e\", \"0x19faaa87edc67b3a03edb38a7690e88bb35f90eaa38556adf4eea04f0a41329b\", \"0x78e2fae5201f6e8b5bcbb7d2cfd9a05b51ee5daac55dc1c67d5bf9496443b124\", \"0x8c473676e83787c7748e9d956e0fd2f316e00dd13771f2cee0367fcf5588f40c\", \"0x6488a7d2ae45fe1f028a2b3d140073efba82aba8db01e9c96d165d515a2b5ead\", \"0xc2baf1c827cfce1eb852561d737b454fe12c7555945ba68017fa6222c93d7f1d\", \"0xc6052c7af376fef1daf078b85913d5b147a68f4c6de75cc9e49399d62a515c92\", \"0xbddc1775065a6ee9386737303af7b9da6ad995c369706fbdcf4a7fcb8cc2cca0\", \"0xd14f617a18014f5bb594e7ca4c999354bc095a6b136a6cd3d097928fe3defd86\", \"0xcfe35d771cfcc111868b623b0885277c74b94f6f5bc544c78c66bc0b5458a524\", \"0x4163b9a6368d125edd8ee3e538f320752b0bd0f113dd79349a72fa3af06c64ab\", \"0x78cb874b15e37a66ae8eee4cf13b1e8af6e63ea72c234067a1023f1f6d989a06\", \"0x13c184ef8cd27a98857dcf2c0a21cefde8808551b4fbc18b1a62b97186662990\", \"0x93584e7b2c0ced44fca221aa0d934cedb1cddb82e0296f041b1934875d505986\", \"0x77865dd539eda3960520bee9cd3e0a8563b9a0ddc854885a1d9dee14b11b3da0\", \"0x8665f83639732248eeba9454540184cc94598e0fe73d126b6e3db31725352f28\", \"0xa322e919186a4b2f2039601db976b005e7b2968e56d75cbf71a77bbe8905292b\", \"0xb995f437374bd1e1a9eb3443bc3d7855f15b4d6ce628051332fe20a3a8cd0e9c\", \"0x75df3f7f39ed5132f0b10ca3d97487eca629128ad37d4e6b5ba391ece4f4702c\", \"0xf59c191e25659b57738ef22ede3d6113091d0425f40ec862de14bf1149fb382e\", \"0xa9588c6ec3513da0ecfdb7967fcf6711ca3566e5ae1b8b60bcc7d54a123f2a07\", \"0x3edc6044a176db79bb6ffccd3f3c950cb5a56b9ba65d4243e9e03cbcf4bbe382\"],    \"Pos\": \"0x1d\"   }  }],  \"Outs\": [{   \"Ar\": null,   \"Asset\": {    \"Tkn\": {     \"Currency\": \"0x000000000000000000000000000000000000000000000000000000005345524f\",     \"Value\": \"1\"    },    \"Tkt\": null   },   \"Memo\": \"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",   \"PKr\": \"0xa8e68b2b7398c188fa431ecf89ce51aa0646c7cc6656cb02de7ee44cc73eec084ea84b8185cfd99d3e5cdb76171a108a35aa858e710dd824175b7f9c35e3f2286ae0e1b5980e64711352822720d083b16472c4cf2226535457cab5b45491f0d8\"  }, {   \"Ar\": null,   \"Asset\": {    \"Tkn\": {     \"Currency\": \"0x000000000000000000000000000000000000000000000000000000005345524f\",     \"Value\": \"7999974999999999999\"    },    \"Tkt\": null   },   \"Memo\": \"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",   \"PKr\": \"0xc200d439f05f4cd23e367abc41c505e0f79990a1eee5613cd2033dee36077511b34214e3e0786e14536d22f0b2d7e0068bec6284b626e46f59f974b21d371020948ab05d5c21304e4c4963375c24736670b63a025c709f0b6b97f04a6c1b8b2c\"  }],  \"Z\": true }"
	param := &txtool.GTxParam{}
	json.Unmarshal([]byte(str), param)
	//param.Z = nil
	ctx, e := SignTx(param)
	fmt.Println(e)
	bytes, _ := json.Marshal(ctx.Tx())
	fmt.Println(string(bytes))
	tx := ctx.Tx()
	_, er := ProveTx(&tx, param)
	fmt.Println(er)

}
