// /public/js/receptionTest.js
(function () {
  console.log("=== Reception ページ自動テスト開始 ===");

  // 初回ロード時に ?reception がなければリダイレクト
  if (!location.search.includes("reception=")) {
    const initReception = {
      receptionTitle: "小劇場",
      askChoices: [
        {
          question: "好きなジャンルは？",
          choices: ["コメディ", "ドラマ", "ホラー"],
          sequence: 0,
        },
      ],
      openTimes: [
        { limitStart: "2025-09-05T18:00", limitEnd: "2025-09-05T22:00" },
      ],
      seats: [
        {
          seatName: "席A-1",
          capacity: 1,
          passcodes: [],
          currentCode: "",
        },
      ],
    };

    const param = encodeURIComponent(JSON.stringify(initReception));
    location.href = `${location.pathname}?reception=${param}`;
    return; // リダイレクトして再実行
  }

  // aliases.value を待つ関数
  async function waitForAliases(timeout = 5000) {
    return new Promise((resolve, reject) => {
      const start = Date.now();
      console.log(window.aliases)
      const timer = setInterval(() => {
        try {
          if (
            window.aliases &&
            Array.isArray(window.aliases.value) &&
            window.aliases.value.length > 0
          ) {
            clearInterval(timer);
            resolve(window.aliases.value);
          }
          if (Date.now() - start > timeout) {
            clearInterval(timer);
            reject(new Error("aliasesが取得できませんでした"));
          }
        } catch (e) {
          // aliases がまだ undefined の場合
        }
      }, 200);
    });
  }

  async function runTest() {
    try {
      console.log("✅ Vueの描画とonMountedの完了待ち…");
      const aliases = await waitForAliases();
      console.log("✅ aliases取得:", aliases);

      const reception = window.reception?.value;
      if (!reception) throw new Error("receptionが存在しません");

      // adminNames に最初の aliasName を設定
      reception.adminNames = [aliases[0].aliasName];
      console.log("✅ adminNamesを設定:", reception.adminNames);

      // openTimes を今日+4日後の09:00–20:00に更新
      const today = new Date();
      today.setDate(today.getDate() + 4);
      const yyyy = today.getFullYear();
      const mm = String(today.getMonth() + 1).padStart(2, "0");
      const dd = String(today.getDate()).padStart(2, "0");

      reception.openTimes = [
        {
          limitStart: `${yyyy}-${mm}-${dd}T09:00`,
          limitEnd: `${yyyy}-${mm}-${dd}T20:00`,
        },
      ];
      console.log("✅ openTimesを設定:", reception.openTimes);

      // submitを発火
      if (typeof window.submit === "function") {
        console.log("✅ submitを実行します…");
        await window.submit();
        console.log("🎉 submit完了!");
      } else {
        console.warn("⚠️ submit関数が見つかりません");
      }
    } catch (err) {
      console.error("❌ テスト失敗:", err.message);
    }
  }

  window.addEventListener("load", runTest);
})();
