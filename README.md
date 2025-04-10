![罷免油雞隊](static/logo.png) 罷免油雞隊

# 用途

- 協助罷免團體機動收件的小工具
- 支援 Google/FB/Line/TG 登入 (不同來源視為不同人)
- 支援多個罷免團隊一起使用 (建議要在地的，避免民眾誤會或交錯件)
- 簡單分為幹部/志工/新手/民眾四種權限
- 志工發佈有空的時間和區域 (`行程`)，聯絡及確認民眾的預約，並依照聯絡確認的內容，在指定時間到指定地點收件
- 民眾選擇行程來掛號 (`預約`)，等志工聯絡討論後，在指定時間去指定地點交件
- 幹部進行管理
  * 聯絡後時間地點喬不攏的，交由幹部處理 (ex 轉給其他志工)
  * 線下進行教育訓練後，給予新手或正式志工權限以進行收件
  * 2FA 有問題時幫忙重設
- 可選用的簡易簽收機制，減少假志工從中作梗的傷害

# 架設

### 設定

- `yugi.example.toml` 有一份完整的設定檔，修改後即可使用
- 每一個選項都可以用 envvar 指定，例如 google 的 oauth client 可以用 `YUGI_AUTH_GOOGLE_CLIENT=asdasdasd` 指定

### Docker

```
docker run -p 8080:8080 -w "$(pwd)/data:/work" --user "$(id -u):$(id -g)" --env-file .env ronmi/yugi
```

### Baremetal

```
yugi serve --config my_config.toml
```

# 設計考量與使用建議

- 志工人力及時間非常有限，所以整個流程的主導者是志工
- 志工與幹部強制使用 2FA 減少盜帳號機會
- 部份匿名化，所有人都使用四字代號
  - 為了管理與協作方便，可以在志工的公開備註記錄方便識別的暱稱、Line id 一類的，方便幹部管理
  - 志工的公開備註只有同一個罷免團體的志工及幹部看得到
  - 這個公開備註會在許多地方使用 (幹部的管理介面、幹部及志工的操作記錄)
  - 民眾沒有備註
- 盡量減少使用 JS，避免舊手機無法使用
- 只是輔助工具，複雜狀況留給人工處理，保留彈性
- 懶得做工程師介面，只弄了一個 subcommand (`addOrg`) 用來初始化帳號 
- 有簡易的樣板檢查 subcommand (`testTemplate`)，不用開 dev server 就能過濾掉一些常見錯誤
- 預約的公開備註，幹部、負責志工及民眾都看得到，所以建議輸入確定的見面時間地點，可以減少爭議
- 秘密備註只有幹部和負責志工看得到，屬於內部溝通使用的欄位
- 建議使用簽收機制，對罷免團體及民眾可以增加一些保障
  * 簽收機制的重點是 `民眾指定一段密語` 以及 `志工記錄收到哪些連署書`，密語只有民眾和幹部看得到
  * 所以就算密語是是由志工代為輸入，事後民眾都能查看確認，民眾更安心的同時，也能提高假志工搗亂的成本
- 四字代碼是透過設定檔中 `randName` 的四個 group 隨機產生，每一個 group 對應一個字，所以同一個 group 裡不能有重複的字，但不同 group 可以重複 (甚至可以四個 group 都用相同的字串)。目前設定的下限是30字，約 81 萬個組合，但程式隨機的方式較簡陋，又可能有網軍干擾，恐怕只能撐住數萬甚至更少民眾使用。 repo 裡的範例設定檔約有 350 萬組合，我個人認為這是實務上較安全的下限。

# License

MPLv2.0

[LOGO 圖片來源](https://commons.wikimedia.org/wiki/File:HK_SKD_%E5%B0%87%E8%BB%8D%E6%BE%B3_TKO_Tseung_Kwan_O_%E5%94%90%E6%98%8E%E8%A1%97_Tong_Ming_Street_%E5%AF%B6%E5%BA%B7%E8%B7%AF_Po_Hong_Road_%E5%AF%8C%E5%BA%B7%E8%8A%B1%E5%9C%92_Beverly_Garden_Shopping_Centre_shop_%E7%87%92%E5%91%B3%E5%BA%97_Siu_Mei_food_%E8%B1%89%E6%B2%B9%E9%9B%9E_soy_sauce_chicken_box_July_2022_Px3_02.jpg)
