{
    _id: '654322',
    api_key: 'abc',
    admin_names: [
        'mik2'
    ],
    join_names: [
        'mik2',
        'ivan1'
    ],
    recept_title: '受付の公開用データ',
    asks: [
        'お名前は？'
    ],
    wait_ratio: 10,
    seats: [
        {
            seat_name: 'table_1',
            capacity: 4,
            passcodes: [
                'AAA',
                'BBB',
                'CCC'
            ],
            current_code: 'CCC'
        },
        {
            seat_name: 'table_2',
            capacity: 4,
            passcodes: [
                'DDD',
                'FFF'
            ]
        }
    ],
    subscription: [
        '{"endpoint":"https://fcm.googleapis.com/fcm/send/cYSrnHSWCtI:APA91bFNaBgx9eRYSZiAGYCxlPAh6e3j8WkFI_1wRdWzqYtpbpxcck76oB9oRFEYiuABy2nWO-O8KsnF9Xo3cPqRhhop7q_BuO0qMYXihDY2kWGhzsdriztgfsei3T6SUnm24eq1YA8Q","expirationTime":null,"keys":{"p256dh":"BBkbQ5jb1u60NSZhxg5fEvsWKkTpW34y-sMxRJP8Oyiscdnj89oZzND7qZEeIEmEUVbgDj5jtkB0d8MTkj7lpoQ","auth":"C8GdBJWx4fAB-HJkvVEzOQ"}}'
    ],
    menus: [
        {
            menu_id: 1,
            menu_name: '味噌ラーメン',
            price: 1000,
            items: [
                2
            ]
        },
        {
            menu_id: 2,
            menu_name: '豚骨ラーメン',
            price: 1100,
            items: [
                4
            ],
            paid_options: [
                {
                    item_id: 7,
                    price: 500
                },
                {
                    item_id: 8,
                    price: 100
                },
                {
                    item_id: 9,
                    price: 400
                }
            ],
            free_options: [
                3,
                4
            ]
        },
        {
            menu_id: 3,
            menu_name: 'お子様ランチ',
            price: 2000,
            items: [
                4
            ],
            paid_options: [
                {
                    item_id: 7,
                    price: 500
                },
                {
                    item_id: 8,
                    price: 200
                }
            ],
            free_options: [
                3,
                4,
                5
            ],
            free_multi_options: [
                5,
                6
            ]
        }
    ],
    item_details: [
        {
            item_id: 1,
            item_name: '豚骨ラーメン',
            choices: [
                [
                    '硬い',
                    '普通',
                    '柔らかい'
                ],
                [
                    '油多め',
                    '普通',
                    '油少なめ'
                ]
            ]
        },
        {
            item_id: 2,
            item_name: 'みそラーメン',
            choices: [
                [
                    '硬い',
                    '普通',
                    '柔らかい'
                ],
                [
                    '油多め',
                    '普通',
                    '油少なめ'
                ]
            ]
        },
        {
            item_id: 3,
            item_name: 'コーラ'
        },
        {
            item_id: 4,
            item_name: 'ジンジャーエール'
        },
        {
            item_id: 10,
            item_name: 'ライス'
        },
        {
            item_id: 11,
            item_name: 'パン'
        },
        {
            item_id: 5,
            item_name: '砂糖'
        },
        {
            item_id: 6,
            item_name: '塩'
        },
        {
            item_id: 7,
            item_name: 'ハム'
        },
        {
            item_id: 8,
            item_name: 'ソーセージ'
        },
        {
            item_id: 9,
            item_name: 'コーヒー',
            choices: [
                [
                    'すぐ',
                    '食後'
                ]
            ]
        }
    ],
    updated_at: ISODate('2025-03-22T01:35:02.961Z')
}