const jwt = require("jsonwebtoken");
const crypto = require("crypto");

const APP_ID = "c0b2a419-9c3c-408f-ab32-671407d6e3ad";
const SECRET_KEY = "E3bVL++TkEprPjK5+DNYdwxuw7yawPv8o1OsBl4pPTU=";

function generateSkyWayToken() {
  const now = Math.floor(Date.now() / 1000);

  const payload = {
    jti: crypto.randomUUID(),
    iat: now,
    exp: now + 60 * 60 * 24 * 365, // 1年

    // 公式 SkyWayAuthToken と同じ構造
    scope: {
      app: {
        id: APP_ID,
        turn: true,
        actions: ["read"],   // 公式コードでは ["read"] のみ

        channels: [
          {
            id: "*",
            name: "*",
            actions: ["write"],

            members: [
              {
                id: "*",
                name: "*",
                actions: ["write"],
                publication: { actions: ["write"] },
                subscription: { actions: ["write"] },
              },
            ],

            sfuBots: [
              {
                actions: ["write"],
                forwardings: [{ actions: ["write"] }],
              },
            ],
          },
        ],
      },
    },
  };

  // HS256 で署名する
  return jwt.sign(payload, SECRET_KEY, { algorithm: "HS256" });
}

console.log(generateSkyWayToken());
