{
    _id: '3eHg',
    admin_names: null,
    books: [
        {
            book_start: '2025-07-04T10:00',
            book_end: '2025-07-04T11:00',
            menu_id: 1
        },
        {
            book_start: '2025-07-04T15:00',
            book_end: '2025-07-04T16:00',
            menu_id: 2
        },
        {
            book_start: '2025-07-05T10:00',
            book_end: '2025-07-05T11:00',
            menu_id: 2,
            created_at: ISODate('2025-07-04T22:21:36.291Z')
        }
    ],
    channel_id: '3eHg',
    facilities: [
        {
            facility_count: 1,
            facility_name: '施術室A'
        },
        {
            facility_count: 1,
            facility_name: '施術室B'
        }
    ],
    seats: [
        {
            seat_name: 'tableA',
            capacity: 4,
            passcodes: [
                {
                    passkey: 'AAA',
                    usage_limit: null,
                    pass_start: '2025-07-01T00:00',
                    pass_end: '2025-08-01T00:00'
                }
            ]
        }
    ],
    join_names: null,
    menus: [
        {
            menu_id: 1,
            menu_name: 'ラーメン',
            price: 800,
            prepaid_price: 0,
            specify_name_flag: false,
            items: [
                101,
                102
            ],
            paid_options: [
                {
                    item_id: 201,
                    price: 100
                },
                {
                    item_id: 202,
                    price: 150
                }
            ],
            free_options: [
                [
                    301,
                    302
                ],
                [
                    303,
                    304
                ]
            ],
            free_multi_options: [
                401,
                402
            ]
        },
        {
            menu_id: 2,
            menu_name: 'チャーハン',
            price: 700,
            prepaid_price: 0,
            specify_name_flag: false,
            items: [
                103
            ]
        }
    ],
    passcodes: [
        {
            passkey: 'AAA',
            usage_limit: null,
            pass_start: '2025-07-01T00:00',
            pass_end: '2025-08-01T00:00'
        }
    ],
    reception_title: 'サンプル受付',
    staff_skills: [
        {
            alias_name: 'スタッフA',
            skills: [
                'カット技術',
                'カラー技術'
            ]
        },
        {
            alias_name: 'スタッフB',
            skills: [
                'カット技術'
            ]
        }
    ],
    updated_at: ISODate('2025-07-05T03:13:13.854Z'),
    work_staffs: [
        {
            alias_name: 'スタッフA',
            work_start: '2025-07-05T09:00',
            work_end: '2025-07-05T17:00',
            seq: 0
        }
    ],
    item_details: [
        {
            item_id: 101,
            item_name: '麺の硬さ',
            choices: [
                [
                    '硬め',
                    '普通',
                    '柔らかめ'
                ]
            ]
        },
        {
            item_id: 102,
            item_name: 'スープの濃さ',
            choices: [
                [
                    '濃いめ',
                    '普通',
                    '薄め'
                ]
            ]
        },
        {
            item_id: 103,
            item_name: 'チャーハンサイズ',
            choices: [
                [
                    '小',
                    '中',
                    '大'
                ]
            ]
        },
        {
            item_id: 201,
            item_name: '味玉'
        },
        {
            item_id: 202,
            item_name: 'チャーシュー追加'
        },
        {
            item_id: 301,
            item_name: 'ネギあり'
        },
        {
            item_id: 302,
            item_name: 'ネギなし'
        },
        {
            item_id: 303,
            item_name: 'ごまあり'
        },
        {
            item_id: 304,
            item_name: 'ごまなし'
        },
        {
            item_id: 401,
            item_name: '紅ショウガ'
        },
        {
            item_id: 402,
            item_name: '辛味ダレ'
        }
    ]
}