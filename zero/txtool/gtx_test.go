package txtool

import (
	"encoding/json"
	"fmt"
	"github.com/dece-cash/go-dece/common"
	"testing"
)

func TestHash(t *testing.T) {
	str := "{\"Gas\":\"0x61a8\",\"GasPrice\":\"0x3b9aca00\",\"Tx\":{\"Desc_O\":{},\"Desc_Z\":{},\"Desc_Pkg\":{},\"Desc_Cmd\":{},\"Tx\":{\"Ins_P\":[{\"Root\":\"0xf7b106944ab25059519a34737ca69602677e2a3990f3df76dae0e1209f655c2e\",\"Nil\":\"0x66e4d272b05d2ce6db32597d9e1ca498205f44108d9d47607d63eb795fb8b051\",\"ASign\":\"0xa87565d07ef40617004d84b3b9c01d9964966b73e51aeb54b8b6af8564c9701ebc829323465fa0eab9145a1a53551ec976dd0a439ca8809cbd8a01cd3f1d7d00\",\"NSign\":\"0x5f6ec704660e3c0885cebd9da6752a0251c1406b3043154cd76cedb14587ed01c6c9ec82ef4d3b438eaf77196bdb3079c515bd4e1f0518e49a60df1f1fb1a4809f5fef3e8fd4b18378b8d6e5a8b2ca911ba6a7994ad46395c04099d3da31428f\"}],\"Ins_P0\":[],\"Ins_C\":[],\"Outs_C\":[{\"AssetCM\":\"0xc35cfaefa805cbf699d4709bd859c3db605266d7b91437aa3eaf02558d9e15a3\",\"PKr\":\"0x9c06afd139faaebf52ce2fa878a3a8634922024241a74aefe1cfa71040ab4591da3bd4e399a12897f08f521d207d1d10960ec1a2765abd7bb1183a779932de25726ba54e9aeb13d9b5ed6684cc4f1d7e02d6395b4a9952381d0bcdc743c8c1d7\",\"RPK\":\"0x84bb7417e4bb006d98bf0b5090462ed6bb97a148e180cdea70b77a10bd45bd0c\",\"EInfo\":\"0xb4310cbce58078bda847647383f7a4e1243dcc307df83dd710e40020f9787a87398544ac35a61ae3da6bf635d011bce6920a038226de74413ee3f9c042c0ca7dd5781772b2bf1b2fd28b3a8416b1dc14b1c80b8805e72868e3fa6048be6ce1ecfa57dc87375cf8e7bc2363e8aeeec1605c2251433d62e65d319edb2fd4c8a83ae008aa77d5a58577a6d79d9a6188bc78c3b65b7b52ca7d2565e0ccc29f4dc415604776ddc0e238c939f55aea670dfa82524bef4bf03ba5d509dc5b39563f1592950065d7c24ddb9367a85f788456edbbd511f4d7be7cbb0e90fe3af9754a9367\"},{\"AssetCM\":\"0x7bf388d5929c20ac0ca7f7496d68946d799b92f66787362eb9a600afa51a10aa\",\"PKr\":\"0x37fffccf6c32303e14670257e87f64cd81781574a04ff5b27fcc18fe0ca8b928f5a62be6d74218c948a4d92862579e1094d2ecf0c1e37e0bc197896524e73c979cce512ef12dddbfc9d019341853207e2396b3aaed7f9d7fd3f31b60e587a867\",\"RPK\":\"0x3d9737be5b9ab7f066b79e17c2d8319dd716df819eab1f8952435172d2c60a1d\",\"EInfo\":\"0x86b0aa3a076a606d116a497ec52e98eb7396462da0fa0732ce85ea115e81a4f006b748e59a65c3f81367eda8b6d13b87a0c1d0ca21654e6ef98228817bbb1c8f89d22a2e21965e1053b6690453dce1a094e9f355c2ee9637f4dc3f2d2a8d3acdb7864752e21d6d362776091754b96bae12b73922ad113d9e7af9e48bbcfd335ad1e11c4091263d12e2fd7cf66d19ec477c2d43a97c8b856a10addac28df26a43b3cf106343a367ba72d0932aff8833f4b648c488cef3b7eabdd1290ce294bbb6c0a9b9c28d695bbdef34f0f3c4745ea419b318d0e2bdbc041ea143c2389d59a5\"}],\"Outs_P\":[]},\"Bcr\":\"0x818abb14d68d167e3f1c322f4d90b170acf494a4e137684d2491801124a385a4\",\"Bsign\":\"0x69d3a1b89ce0fca3dc4af13b09bb563401963796cb3a0021ddcc26ac7de8008bff93e3ac9336aa3f9e711810b09ab0f4cd479d5747d6e48ef289a3e4f4278101\",\"Ehash\":\"0x59bd488f5e20e9b883fb510f5cf33ead2105f39adaec025bf9218e4e16a2e2b0\",\"From\":\"0x37fffccf6c32303e14670257e87f64cd81781574a04ff5b27fcc18fe0ca8b928f5a62be6d74218c948a4d92862579e1094d2ecf0c1e37e0bc197896524e73c979cce512ef12dddbfc9d019341853207e2396b3aaed7f9d7fd3f31b60e587a867\",\"Fee\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":\"25000000000000\"},\"Sign\":\"0xfa4be5a89c70b7663995af09583ab067c5e1685b160a7a941bff9e531e92d02f88efca9ac2923065106c9dd0a195f0a6699f5079e9292ce8f60888d694386403\"},\"Hash\":\"0xb9793158922500e8330c2c87bd3099d761724a47a1edbb9158bb89a87bef11e8\"}";
	tx := GTx{};
	json.Unmarshal([]byte(str), &tx);
	ret := tx.Tx.ToHash()
	fmt.Println(common.Bytes2Hex(ret[:]));
	tx1_hash := tx.Tx.Tx1_Hash()
	fmt.Println("tx1_hash", common.Bytes2Hex(tx1_hash[:]));
	tx1_tohash := tx.Tx.Tx.ToHash()
	fmt.Println("tx1_tohash", common.Bytes2Hex(tx1_tohash[:]));
	// bytes, _ := json.Marshal(tx)
	// fmt.Println(string(bytes));
}

func TestJson()  {
	// str := "{\"Gas\":90000,\"GasPrice\":1000000000,\"Fee\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":\"90000000000000\"},\"From\":{\"SKr\":\"0x8e0d8402367784517c1a7f3f11940747f519d5dc33636ad2af373ee6b44877006c4d2a8607922277d6a0635b2ff6fa25d7db06ed65619507169e0ee7089203030000000000000000000000000000000000000000000000000000000000000000\",\"PKr\":\"0xed75a97a82f8266115e967734747de0db062b467563116fbfd73d909c9b464aff47ca967203cafc7aaf4f89d0282bbdeebcda08663cdf9d7c57732a396a1fa068289fa3f5f27c44a936e08b734de8b4af103c952589a917539077bb901431bc1\"},\"Ins\":[{\"SKr\":\"0x8e0d8402367784517c1a7f3f11940747f519d5dc33636ad2af373ee6b44877006c4d2a8607922277d6a0635b2ff6fa25d7db06ed65619507169e0ee7089203030000000000000000000000000000000000000000000000000000000000000000\",\"Out\":{\"Root\":\"0x683dfbbba9e686db3f2ca23f2cb993cca9427bb0e1e2ce510f3bad9293c24a2f\",\"State\":{\"OS\":{\"Index\":21,\"Out_O\":null,\"Out_Z\":null,\"Out_P\":{\"PKr\":\"0x385f1b179fb44e9f804c3e8a2e742f832207f3e3330e082c0abbd1b3c905eb2ca2e7f8be409341ec2259572ca3b8193d19b929411270233464c5f9fbab28cd2ebfe1442e0a69888e69d20abc1ddca7bd619a5ddec66184cc106f10f6e7172ef0\",\"Asset\":{\"Tkn\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":\"65000000000000\"},\"Tkt\":null},\"Memo\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"},\"Out_C\":null,\"OutCM\":null,\"RootCM\":\"0xe2e35de2e89d05abb6988f8a6418dee37cac0ca34b0a969e2a68846a26311c9a\"},\"TxHash\":\"0xd2c9507cfda88ce89542798a743a343d9d6101d163273e28242fc4ab0f659c87\",\"Num\":79}},\"Witness\":{\"Pos\":\"0x15\",\"Paths\":[\"0x568c4229a69a833b0de0a0716a2de1c45cd16da96bd0383b09e07f96d5cf1683\",\"0x0100000000000000000000000000000000000000000000000000000000000000\",\"0xff6a7f715e73e79878d4c5ee6ea1133b056f078bef296346e01605dccdeb4d97\",\"0x130adf7ac88bef31232230e2954ce96f35945f640c7517930ab14d89a995bc1c\",\"0xf507f7f38d8c1cd206d898e1bde0285f81b39381a7275bd53732b0cca8278c2c\",\"0x9cda5493957ef78fa117f61252eb9b79e835906dd62f4f230964b4c2b790b892\",\"0xff3bd1141de02f0c80b435835dbb245fd5f47295e94df8935269940ae3940209\",\"0xf57f2632fe188f663c4bd0215aab05170aeb2abebadd28132ef279c5b49f79ae\",\"0xc0f73a47ede0ce8d04cc301e0ecd44148fbf887068b153e297550690fe67061f\",\"0x75eb8ed603b7e4ef3f5ad521f318b08030b25ea492ff84bc20773ed75d88570b\",\"0x9bdd82fcd971d4159e17a321242575fc15d6c56151edccc9f312a8d7e75edb2a\",\"0x50813d276f42fc0a25719f10ef573feef34ce47e2cdcdfd8ae0fa17dce70e698\",\"0xcbd115a1efe853defd7fc120b08606fa7f389c3f9425b747950682c1e09deb12\",\"0x80c0d8b03bff5f553da3fc641866ac3afbbf310908569daddf5e0e7092ad62a0\",\"0x7865c55ce8f7cb355aca5f593426d38b5f82bf0b9f4d19784d75ab498b3ca8ac\",\"0x2ceadea63565bfa72235374aa70c6ff9146456dab2a1febb1b83df2bb0b5bf90\",\"0x3686698d6e508112eb93891d48c7b50e12d54914bfdd437076ceaa60ef696e88\",\"0x311565e80def136b1a7838d09a2f1f0bfd7b7ac2a6d04d4cb6cb2efd4b33b011\",\"0x9f3b87c8faa34d9104e865b38b3cb87bc5ff71cb0a36a23b71723fbfd31fa72a\",\"0xdbdbab4b98abd0fb19465930710c700bfb07cede792139f1b4a49f340a161f24\",\"0x68b4a158325154aee284f9aec951c0da13fe4b44c9188973644a7b68f78735ac\",\"0xfb9cbf55b5bb63b8e420be52a51ffcf8bbbe878afc696ca70f3f4f2f3358260c\",\"0x31f7aa53943b9bfe0ccc6f8e25479952f2b446ffa731ef90e45a395b2b4a6e09\",\"0x447bdf93bb335d8b268dc74569477a5717197fe538795acdf6ab63fb3d848a98\",\"0x79808b4393da03ae5d0b064b532972d235b571fbc7d5900fe745834385c6981c\",\"0x042415ac74349b1eb318c113474f3b2373fcf4915f9224208f8a70224c10ac03\",\"0xf3519d004458238b7277d6cae47c966cc1b2131bbc14bb6ad8c95d4b1cb75425\",\"0xbc656cd9fd825411e4f2a1da79073c1bb4083432a6aebb80947be38a94cdba80\",\"0x24378894a012b4e87b04e8d4613b81d1795d87089f83c86f6c42cd8c3c67898e\"],\"Anchor\":\"0x683dfbbba9e686db3f2ca23f2cb993cca9427bb0e1e2ce510f3bad9293c24a2f\"},\"A\":null,\"Ar\":null},{\"SKr\":\"0x8e0d8402367784517c1a7f3f11940747f519d5dc33636ad2af373ee6b44877006c4d2a8607922277d6a0635b2ff6fa25d7db06ed65619507169e0ee7089203030000000000000000000000000000000000000000000000000000000000000000\",\"Out\":{\"Root\":\"0xb146c4181ae841c64f5eb20e090e01cd580a28b6774d7c10226dfe603bd3fb22\",\"State\":{\"OS\":{\"Index\":20,\"Out_O\":null,\"Out_Z\":null,\"Out_P\":null,\"Out_C\":{\"PKr\":\"0x385f1b179fb44e9f804c3e8a2e742f832207f3e3330e082c0abbd1b3c905eb2ca2e7f8be409341ec2259572ca3b8193d19b929411270233464c5f9fbab28cd2ebfe1442e0a69888e69d20abc1ddca7bd619a5ddec66184cc106f10f6e7172ef0\",\"AssetCM\":\"0x37967302a56a32104a73edda03c03bc65929e84324080d72bc093ee057eac196\",\"RPK\":\"0x37daffa4de3416dd57e1a36256224fd63050f9714593ff791a2a0747727dcc96\",\"EInfo\":\"0xbacbf563267e46729761d0909fca5cc322a01c5c131811daa50ecec3f7fad0830e44a1c6435ae6bd4715c3a3e8e6a1b100832c8350684e83d7a2c64d623fe717809df12528f58bc1ae0ddfa8aeb029a8ffbc910d8a2e3d43760eecb0344c2af1f31b0256bfbaaaeb3960f85665dc566f6ab4941051baab11951bd169835356326219babb70d4ae466ca11661cdefdab339389d37e763531b3de9754bbcb391eaf68923c3feaa0a76c86510ec0a54fc78ce005e4f8f478a29e36063651e239bd2600a5ea444cf1b8381294a099429e0db319c14d8b444a211cb1a1d9a561d2e08\",\"Proof\":\"0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"},\"OutCM\":null,\"RootCM\":\"0x568c4229a69a833b0de0a0716a2de1c45cd16da96bd0383b09e07f96d5cf1683\"},\"TxHash\":\"0xd2c9507cfda88ce89542798a743a343d9d6101d163273e28242fc4ab0f659c87\",\"Num\":79}},\"Witness\":{\"Pos\":\"0x14\",\"Paths\":[\"0xe2e35de2e89d05abb6988f8a6418dee37cac0ca34b0a969e2a68846a26311c9a\",\"0x0100000000000000000000000000000000000000000000000000000000000000\",\"0xff6a7f715e73e79878d4c5ee6ea1133b056f078bef296346e01605dccdeb4d97\",\"0x130adf7ac88bef31232230e2954ce96f35945f640c7517930ab14d89a995bc1c\",\"0xf507f7f38d8c1cd206d898e1bde0285f81b39381a7275bd53732b0cca8278c2c\",\"0x9cda5493957ef78fa117f61252eb9b79e835906dd62f4f230964b4c2b790b892\",\"0xff3bd1141de02f0c80b435835dbb245fd5f47295e94df8935269940ae3940209\",\"0xf57f2632fe188f663c4bd0215aab05170aeb2abebadd28132ef279c5b49f79ae\",\"0xc0f73a47ede0ce8d04cc301e0ecd44148fbf887068b153e297550690fe67061f\",\"0x75eb8ed603b7e4ef3f5ad521f318b08030b25ea492ff84bc20773ed75d88570b\",\"0x9bdd82fcd971d4159e17a321242575fc15d6c56151edccc9f312a8d7e75edb2a\",\"0x50813d276f42fc0a25719f10ef573feef34ce47e2cdcdfd8ae0fa17dce70e698\",\"0xcbd115a1efe853defd7fc120b08606fa7f389c3f9425b747950682c1e09deb12\",\"0x80c0d8b03bff5f553da3fc641866ac3afbbf310908569daddf5e0e7092ad62a0\",\"0x7865c55ce8f7cb355aca5f593426d38b5f82bf0b9f4d19784d75ab498b3ca8ac\",\"0x2ceadea63565bfa72235374aa70c6ff9146456dab2a1febb1b83df2bb0b5bf90\",\"0x3686698d6e508112eb93891d48c7b50e12d54914bfdd437076ceaa60ef696e88\",\"0x311565e80def136b1a7838d09a2f1f0bfd7b7ac2a6d04d4cb6cb2efd4b33b011\",\"0x9f3b87c8faa34d9104e865b38b3cb87bc5ff71cb0a36a23b71723fbfd31fa72a\",\"0xdbdbab4b98abd0fb19465930710c700bfb07cede792139f1b4a49f340a161f24\",\"0x68b4a158325154aee284f9aec951c0da13fe4b44c9188973644a7b68f78735ac\",\"0xfb9cbf55b5bb63b8e420be52a51ffcf8bbbe878afc696ca70f3f4f2f3358260c\",\"0x31f7aa53943b9bfe0ccc6f8e25479952f2b446ffa731ef90e45a395b2b4a6e09\",\"0x447bdf93bb335d8b268dc74569477a5717197fe538795acdf6ab63fb3d848a98\",\"0x79808b4393da03ae5d0b064b532972d235b571fbc7d5900fe745834385c6981c\",\"0x042415ac74349b1eb318c113474f3b2373fcf4915f9224208f8a70224c10ac03\",\"0xf3519d004458238b7277d6cae47c966cc1b2131bbc14bb6ad8c95d4b1cb75425\",\"0xbc656cd9fd825411e4f2a1da79073c1bb4083432a6aebb80947be38a94cdba80\",\"0x24378894a012b4e87b04e8d4613b81d1795d87089f83c86f6c42cd8c3c67898e\"],\"Anchor\":\"0x683dfbbba9e686db3f2ca23f2cb993cca9427bb0e1e2ce510f3bad9293c24a2f\"},\"A\":null,\"Ar\":null}],\"Outs\":[{\"PKr\":\"0xa83e205ebf864e5ed78bce2dd3766253cc553324a8134ae517aefedc43700793366d8a85742a26a5d02498f6ceaa7aa4778bb6c029b39f952e7e88bfd319b7201721c261629c0677a3aaa743b85210a4f5e9f9305debb9bd6c444fdb63edcbc6\",\"Asset\":{\"Tkn\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":\"589925000000\"},\"Tkt\":null},\"Memo\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"Ar\":null},{\"PKr\":\"0xed75a97a82f8266115e967734747de0db062b467563116fbfd73d909c9b464aff47ca967203cafc7aaf4f89d0282bbdeebcda08663cdf9d7c57732a396a1fa068289fa3f5f27c44a936e08b734de8b4af103c952589a917539077bb901431bc1\",\"Asset\":{\"Tkn\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":\"9884410075000000\"},\"Tkt\":null},\"Memo\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"Ar\":null}],\"Cmds\":{\"BuyShare\":null,\"RegistPool\":null,\"ClosePool\":null,\"Contract\":null,\"PkgCreate\":null,\"PkgTransfer\":null,\"PkgClose\":null}}"

}