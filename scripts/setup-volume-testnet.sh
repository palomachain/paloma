#!/bin/bash
set -euo pipefail
set -x

if ! which jq > /dev/null; then
  echo 'command jq not found, please install jq'
  exit 1
fi

if [[ -z "${CHAIN_ID:-}" ]]; then
  echo 'CHAIN_ID required'
  exit 1
fi

if [[ -z "${MNEMONIC:-}" ]]; then
  echo 'MNEMONIC required'
  exit 1
fi

jq-i() {
  edit="$1"
  f="$2"
  jq "$edit" "$f" > "${f}.tmp"
  mv "${f}.tmp" "$f"
}

palomad init birdlady --chain-id "$CHAIN_ID"

pushd ~/.paloma/config/
sed -i 's/^keyring-backend = ".*"/keyring-backend = "test"/' client.toml
sed -i 's/^minimum-gas-prices = ".*"/minimum-gas-prices = "0.001ugrain"/' app.toml
sed -i 's/^laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq-i ".chain_id = \"${CHAIN_ID}\"" genesis.json
popd

GR=000000ugrain
KGR="000${GR}"
MGR="000000${GR}"

INIT_AMOUNT="5${MGR}"
INIT_VALIDATION_AMOUNT="10${KGR}"
GENESIS_AMOUNT="1${MGR}"
FAUCET_AMOUNT="3${MGR}"

name="splat"
echo "$MNEMONIC" | palomad keys add "$name" --recover
address="$(palomad keys show "$name" -a)"

palomad add-genesis-account "$address" "$INIT_AMOUNT"
palomad gentx "$name" "$INIT_VALIDATION_AMOUNT" --chain-id "$CHAIN_ID"

init() {
  name="$1"
  address="$2"
  amount="${3:-"$GENESIS_AMOUNT"}"

  palomad add-genesis-account "$address" "$amount"
}

init faucet paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn "$FAUCET_AMOUNT"
init chase paloma1nty4gn8k2nrewy26fm62v03322fxgpq0hxssn6
init jason paloma1mre80u0mmsdpf3l2shre9g4sh7kp9lxu5gtlql
init birdpoop2 paloma1k4hfe8cqdzy6j0t7sppujhrplhp8k5tglf8v46
init steven paloma1arwskydtxc8w6jw0pa227yslmzrwllspxrlaga
init zhibai paloma1fv2t7yf0dvlc4yz888qfamyys35l0rf0xdwwfl
init william paloma17esj9dnjhnpezfqmpwdwg2yldnc3udrdv5e76w

init  1 paloma1hj4lxqp05ntjl6ezu3qn9eyedp5p4fpdwu8gxj "$INIT_VALIDATION_AMOUNT"
init  2 paloma1d3v3jh6l2r23y9kgzdrahx0ev8ez0g8qapsfxs "$INIT_VALIDATION_AMOUNT"
init  3 paloma16jhsvx4zrukkjqd9akfawx2tzduyx4uxgtjj42 "$INIT_VALIDATION_AMOUNT"
init  4 paloma15gvyk43x406v7kcd4rff5qfutqmcnpj3p4ea9g "$INIT_VALIDATION_AMOUNT"
init  5 paloma1z9fgzh7mzqgu33pdkxw0dqmqgm9l8exj4nggcp "$INIT_VALIDATION_AMOUNT"
init  6 paloma1ljg6ed0pzc3xpqtareyfp6h4fpngs7nw0nnu2j "$INIT_VALIDATION_AMOUNT"
init  7 paloma13uslh0y22ffnndyr3x30wqd8a6peqh255hkzrz "$INIT_VALIDATION_AMOUNT"
init  8 paloma1qf7np0rp3qutvn08vc6qz0y7ffc02cncpam3a7 "$INIT_VALIDATION_AMOUNT"
init  9 paloma10y227j9d09pckexy32v2gckerj9a0kcepc7zsh "$INIT_VALIDATION_AMOUNT"
init 10	paloma1swa5kcf9cl5dx2ypx0c5r9e5qdfnzp9wq59z2l "$INIT_VALIDATION_AMOUNT"
init 12	paloma1espezfvhuagml6g0flw8jj6ajs2wud4cy7wgdu "$INIT_VALIDATION_AMOUNT"
init 13	paloma1tdw23fpnxh2uk3djtteh7eaydymrfgnak3paq3 "$INIT_VALIDATION_AMOUNT"
init 14	paloma1pmx8m7glpw4pckwzrfflqp4427ynl32wam45z4 "$INIT_VALIDATION_AMOUNT"
init 15	paloma1ylmkqmwct392ytu52tncpeyz4dskklx56cd3rf "$INIT_VALIDATION_AMOUNT"
init 16	paloma16lez38lgsgu34ka2gq8yee8a862zpgs4rt52xs "$INIT_VALIDATION_AMOUNT"
init 17	paloma1am3k7czusdcewv55nhaugu2drn38af9449yxzj "$INIT_VALIDATION_AMOUNT"
init 18	paloma1pdd0nmuj0xfwhfyt7h3wkx9zgjvs3hzlehakd5 "$INIT_VALIDATION_AMOUNT"
init 19	paloma1kludne80z0tcq9t7j9fqa630fechsjhxhpafac "$INIT_VALIDATION_AMOUNT"
init 20	paloma14jux7zw5qd9gdrapypgndwysrmn29gx2nzsf70 "$INIT_VALIDATION_AMOUNT"

init 21 paloma19svt6tkvnu9wcjfwz2daglnxtm0frcavpzeyk2 "$INIT_VALIDATION_AMOUNT"
init 22 paloma1my3gpyx7sdx7wn4rd0hmng60q9jhykxhfp7jqy "$INIT_VALIDATION_AMOUNT"
init 23 paloma1j2zvqhxqycxlxj3stnmun8060wfhfw57up40dc "$INIT_VALIDATION_AMOUNT"
init 24 paloma1crwvzzz2jksyjva3rzewdjk2s4jlxp682a5vrf "$INIT_VALIDATION_AMOUNT"
init 25 paloma13cfxrvldlpxdhn8mq9ydm3syyshddruzn45mvh "$INIT_VALIDATION_AMOUNT"
init 26 paloma1ksxkxz686m8jw0rvpspdlgrd5557jy84c66lcj "$INIT_VALIDATION_AMOUNT"
init 27 paloma1xmm7rf06d3d8l5zmj2jmcjrfftf5kctn9602gt "$INIT_VALIDATION_AMOUNT"
init 28 paloma174l5um5rahquqxlvchyhfeveuue7j8faw7atq7 "$INIT_VALIDATION_AMOUNT"
init 29 paloma1sz3zjcyq8cd3e2k9kx6d44s8wnavfsy2gpf0y2 "$INIT_VALIDATION_AMOUNT"
init 30 paloma18q0s0zj4jks7d9vflzt7nzrdjhyry007fmzr57 "$INIT_VALIDATION_AMOUNT"
init 31 paloma1m986m3a8ag2xaezdda2vjfg7uwecjezetr839f "$INIT_VALIDATION_AMOUNT"
init 32 paloma1d0py2ucz2jrz69ca5l3pd4dfrmv64r24wu00ja "$INIT_VALIDATION_AMOUNT"
init 33 paloma1p4r04ek0jnngegmw76rpx66xu9ep0kuzrge6jh "$INIT_VALIDATION_AMOUNT"
init 34 paloma1n3lnqk3mf5seau5rcnf0jysraq2ah73h5vpqed "$INIT_VALIDATION_AMOUNT"
init 35 paloma1nlwzcegm7kxg6rpmkqmpl3afvapkyp4ml0f5cm "$INIT_VALIDATION_AMOUNT"
init 36 paloma1fj23wrmsemxqe8qut84vphqj4y0g685s8wh0vl "$INIT_VALIDATION_AMOUNT"
init 37 paloma1mcatnt4kg7ag8tdgeynzuf6d2qs225tq5fw42e "$INIT_VALIDATION_AMOUNT"
init 38 paloma1n7ypqlhwtygn8aucp6m0wwxn0umeljnlrncpgg "$INIT_VALIDATION_AMOUNT"
init 39 paloma1s2jzpdfedxqfmttl55f9yvhmpgj738t2qkuaql "$INIT_VALIDATION_AMOUNT"
init 40 paloma1jr353upa2jcg2af20lkq5ch60npk4a3wmcut9q "$INIT_VALIDATION_AMOUNT"
init 41 paloma16j8tzk5kg2kxqu54gwwtjdktst8uy02lchj83c "$INIT_VALIDATION_AMOUNT"
init 42 paloma1c8uxkkz00qn97wsp9um8y3a8wlmrv3c5629jrn "$INIT_VALIDATION_AMOUNT"
init 43 paloma1xx9pqn0ejf4gv48an498kjtv0a0s6c72c6zqs3 "$INIT_VALIDATION_AMOUNT"
init 44 paloma19dd8gaemmrn6q7clvn5xkskccdfwedk5j98zca "$INIT_VALIDATION_AMOUNT"
init 45 paloma18p8gsyvtlmnc4j5ruu5803ea6k7j2akr3sufff "$INIT_VALIDATION_AMOUNT"
init 46 paloma12tyjlznq3n6w7nzvpct3mullzqt3kjp5rjk52v "$INIT_VALIDATION_AMOUNT"
init 47 paloma1k4xsweufff7p6duhcz5468rjxegrcplj7rnyng "$INIT_VALIDATION_AMOUNT"
init 48 paloma1tsu8nthuspe4zlkejtj3v27rtq8qz7q62hx7ae "$INIT_VALIDATION_AMOUNT"
init 49 paloma15u4u7yz8v8p6edytx9d4t4tm6hy3rvwsqw88nj "$INIT_VALIDATION_AMOUNT"
init 50 paloma1vglwq68encr73tlvcqf94m82j0gt2lefwpz7xw "$INIT_VALIDATION_AMOUNT"
init 51 paloma1x007zwjwmspkweeelu0uvfhqtf8d6wdj49tsh9 "$INIT_VALIDATION_AMOUNT"
init 52 paloma18dyx5c70mtl6p0trgx6a2eg46f44q093j568y0 "$INIT_VALIDATION_AMOUNT"
init 53 paloma1y85t2eemc2hhw57jd5p4jetmxq724npa4xe24j "$INIT_VALIDATION_AMOUNT"
init 54 paloma1ug059p3wyxrxkj36hmntmvjcl77z9dse4fh5dj "$INIT_VALIDATION_AMOUNT"
init 55 paloma1tw69r8yh2s8zy8rs7uufpy2f29zdfaer9jeyef "$INIT_VALIDATION_AMOUNT"
init 56 paloma1u5xmuzrket78a6lw2shm06sluxcfhsrr08j6h0 "$INIT_VALIDATION_AMOUNT"
init 57 paloma13jql7e3yx4n8xjj26mnt9jzytajpjyw6ggsh5k "$INIT_VALIDATION_AMOUNT"
init 58 paloma12gxarx0lk2atvpwsvgsxkucglhu6gh4675c2tr "$INIT_VALIDATION_AMOUNT"
init 59 paloma1q90vccjd2ewhngpkjc9rudwyz0pt9fxhej3ust "$INIT_VALIDATION_AMOUNT"
init 60 paloma1cj9ltsexy0sm6kyrfqnwk0gqaje2l49xpj2x0x "$INIT_VALIDATION_AMOUNT"
init 61 paloma1flgt0un95wlj2j6ws4xfsw5vfefmvzm98rx60l "$INIT_VALIDATION_AMOUNT"
init 62 paloma1h2adfqwras5q8p3hfgslfevj2ml7zs9srclv3s "$INIT_VALIDATION_AMOUNT"
init 63 paloma1x3sngn5exusll2fhly6vh7el7v2t4lfxw06lv6 "$INIT_VALIDATION_AMOUNT"
init 64 paloma17jzj4m5d6uaqh5s7slf9xhmwy4mkwxm9n0ymff "$INIT_VALIDATION_AMOUNT"
init 65 paloma1cs6trg6chgcw0t8dzjx9tup7emctxxtwehg6kq "$INIT_VALIDATION_AMOUNT"
init 66 paloma1zhhwr8gk8pqf9p9eamxnqqstmtszjazwywr285 "$INIT_VALIDATION_AMOUNT"
init 67 paloma13u46lqvkuk54h9f068u2vvmucc7lsw22ycdu47 "$INIT_VALIDATION_AMOUNT"
init 68 paloma14crw5q2e8fh5wrx6jwa5tfy89er9r6phc32nlv "$INIT_VALIDATION_AMOUNT"
init 69 paloma175cy5yfmwx9ffa65cdfyzesqfzvxpwyhmfupq4 "$INIT_VALIDATION_AMOUNT"
init 70 paloma1xaelfvaltqr3g7lurvrxtuh4krhmtgw4jxw94z "$INIT_VALIDATION_AMOUNT"
init 71 paloma1z0duu30tevz8jj323rhetsc7q99cq88e3vr7dx "$INIT_VALIDATION_AMOUNT"
init 72 paloma1848ywe2gnd23faedwzhzmxqgspp269fww39h77 "$INIT_VALIDATION_AMOUNT"
init 73 paloma17kdfltcu45llx54ue0fvj8m4z9gd3ps4eg7v2n "$INIT_VALIDATION_AMOUNT"
init 74 paloma10gsgzpryhkz34u6pdm97kjla8hkx2pwl49ncnu "$INIT_VALIDATION_AMOUNT"
init 75 paloma1urx2ut6s33ngd8d4q7s969pe7w27ca0xgrtcf5 "$INIT_VALIDATION_AMOUNT"
init 76 paloma1q9camqflltxgj3jwl557p6q4s94ycdsu64dypw "$INIT_VALIDATION_AMOUNT"
init 77 paloma13s72wfjtx535uejex2ajk78dc9uwy3lml5mvdy "$INIT_VALIDATION_AMOUNT"
init 78 paloma1jwwd4xdpn70rwdr6qc2ma772s3ksrrrhmukmfh "$INIT_VALIDATION_AMOUNT"
init 79 paloma1p0u43r33se80h9jjc33tqn4tyj9fuuagf0v92p "$INIT_VALIDATION_AMOUNT"
init 80 paloma1zz9fnrcpse9fjrq2sjg6rtxhv9z3uyeswyxejx "$INIT_VALIDATION_AMOUNT"
init 81 paloma1mjtdmmw6lzv29sne7uvm4klvry3twrpp6y6zkj "$INIT_VALIDATION_AMOUNT"
init 82 paloma14xpt92spnm96tc7xmdm3ufu7z44ymnfqh0zpt5 "$INIT_VALIDATION_AMOUNT"
init 83 paloma159z2evzvypm4gcdcneyn0y5mp62n2hxnvpee03 "$INIT_VALIDATION_AMOUNT"
init 84 paloma1q3luz08p8ayjsyygkwl9td0fsn0h907dc3uztp "$INIT_VALIDATION_AMOUNT"
init 85 paloma1pan8sfwsjqlm4gqzd2wsx28ugcnn9m69nv5nq8 "$INIT_VALIDATION_AMOUNT"
init 86 paloma1nevk879ka6uwljg2eth6zpdt4yvfm8kfplmntd "$INIT_VALIDATION_AMOUNT"
init 87 paloma1es6tyl8f6k6phx47vq66le60mh629t3xr92le0 "$INIT_VALIDATION_AMOUNT"
init 88 paloma1elhz6uuxgjjfqd78pgcaz65mxyuvyxhevf8ryt "$INIT_VALIDATION_AMOUNT"
init 89 paloma1npxs5nr96twkn8fp00pq63dv52c2mqn6e6s6le "$INIT_VALIDATION_AMOUNT"
init 90 paloma18tz9pqy2aaadax29xnszyvp805cf0fdpxtmz4h "$INIT_VALIDATION_AMOUNT"
init 91 paloma1kqufqd069ewrst92augwrywaagkwyrjs0jkkez "$INIT_VALIDATION_AMOUNT"
init 92 paloma1npwku4dmlnfwx9vqjnxzpfvdyxtcqvh3cqcl5x "$INIT_VALIDATION_AMOUNT"
init 93 paloma10jvz4jgd84tzcqrrumxmas9pu50js7xfcjf5en "$INIT_VALIDATION_AMOUNT"
init 94 paloma1aec790r8v55a5kntfmmver0zwvekptggt7n6j2 "$INIT_VALIDATION_AMOUNT"
init 95 paloma1khpxtmgnwyeeqvgqgw6kjghmer858kd2nn9sp5 "$INIT_VALIDATION_AMOUNT"
init 96 paloma13mqzh0964cyz8jh0ycr4xrz86gvjsxjpevmrzt "$INIT_VALIDATION_AMOUNT"
init 97 paloma12lajlnny0y0dcdugwlyjk5lcqlk8jsy95pfpj4 "$INIT_VALIDATION_AMOUNT"
init 98 paloma1flg0kc4zcungh3lfnw52a9ahwl0dt7ch2cz9nl "$INIT_VALIDATION_AMOUNT"
init 99 paloma1sysfu6jw5q7za5t5ddhkjm7uvanvy6nm5ac5j8 "$INIT_VALIDATION_AMOUNT"
init 100 paloma1ftpvuxvnuspg9awv4y90h4wqswanadtg5w4qu5 "$INIT_VALIDATION_AMOUNT"
init 101 paloma1v7v3trvehram4363e3dv0w0av5wrwvqqtt4u7z "$INIT_VALIDATION_AMOUNT"
init 102 paloma10ltyzh8ah0y64d42uuatfhumyr84778hywc5e4 "$INIT_VALIDATION_AMOUNT"
init 103 paloma12yv9vmlgwu0p5vut8gc3tsujg7lf9g4xwpxqkp "$INIT_VALIDATION_AMOUNT"
init 104 paloma1f807av4wlphx95dmwy9ayrfv73r4yt7qq0f9g5 "$INIT_VALIDATION_AMOUNT"
init 105 paloma19j2az2h4vmgt5d6z70d9y2zdpwhjpxu89mwkx7 "$INIT_VALIDATION_AMOUNT"
init 106 paloma1p39v659x9kqvd28jpasq3lkdprymhcy4fz9t4z "$INIT_VALIDATION_AMOUNT"
init 107 paloma1q6kj48j9087drhyg5z5jmw5m2nslwzff4djktx "$INIT_VALIDATION_AMOUNT"
init 108 paloma15ejzqg4rtsngnk6zkp58tej3ec9qfm5fdw7vc2 "$INIT_VALIDATION_AMOUNT"

palomad collect-gentxs
