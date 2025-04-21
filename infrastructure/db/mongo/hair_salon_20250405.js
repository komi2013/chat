{
    _id: '123456',
    admin_names: [
        'mik2'
    ],
    join_names: [
        'ivan1',
        'mik2'
    ],
    reception_title: 'サロンの公開用予約リンク',
    asks: [
        '名前は？'
    ],
    ask_choices: [
        [
            '性別は？',
            '男',
            '女',
            'その他'
        ]
    ],
    ask_multi_choices: [
        [
            '何歳ですか？',
            '~ 15',
            '16 ~ 18',
            '19 ~ 25',
            '26 ~ 35',
            '35 ~'
        ]
    ],
    facilities: [
        {
            facility_count: 1,
            facility_name: 'perm'
        }
    ],
    shifts: [
        {
            alias_names: [
                'mik2',
                'ivan1'
            ],
            shift_start: '2025-04-15T10:00',
            shift_end: '2025-04-15T23:00',
            open: 1,
            role: 'stylist',
            fix: true
        },
        {
            alias_names: [
                'ivan1',
                'mik2',
                'mik3'
            ],
            shift_start: '2025-04-16T15:00',
            shift_end: '2025-04-16T20:00',
            open: 2,
            role: 'stylist',
            fix: true
        },
        {
            alias_names: [
                'ivan1',
                'mik2'
            ],
            shift_start: '2025-04-17T15:00',
            shift_end: '2025-04-17T20:00',
            open: 1,
            role: 'helper'
        }
    ],
    open_times: [
        {
            limit_start: '2025-03-19T10:00',
            limit_end: '2025-03-19T20:00'
        },
        {
            limit_start: '2025-03-14T15:00',
            limit_end: '2025-03-14T20:00'
        }
    ],
    menus: [
        {
            menu_id: 1,
            menu_name: 'パーマ',
            price: 10000,
            need_skill: 'perm',
            need_facility: 'perm',
            spend_minute: 60
        },
        {
            menu_id: 2,
            menu_name: 'カット',
            need_skill: 'cut',
            price: 3000,
            spend_minute: 60
        }
    ],
    skills: [
        'cut',
        'perm'
    ],
    staff_skills: [
        {
            alias_name: 'ivan1',
            skills: [
                'perm',
                'cut'
            ]
        },
        {
            alias_name: 'mik2',
            skills: [
                'perm'
            ]
        }
    ],
    work_staff_need: true,
    books: [
        {
            book_start: '2025-03-25T10:30',
            book_end: '2025-03-25T11:00',
            menu_id: 2
        },
        {
            book_start: '2025-03-25T16:00',
            book_end: '2025-03-25T17:00',
            menu_id: 1
        },
        {
            book_start: '2025-03-25T16:30',
            book_end: '2025-03-25T17:30',
            menu_id: 1
        },
        {
            book_start: '2025-03-25T18:20',
            book_end: '2025-03-25T19:00',
            answers: [
                'aaa',
                '女',
                '16 ~ 18'
            ],
            menu_id: 1
        },
        {
            book_start: '2025-03-25T15:00',
            book_end: '2025-03-25T16:00',
            answers: [
                'おおお',
                '女',
                '16 ~ 18'
            ],
            menu_id: 1,
            created_at: ISODate('2025-03-22T01:35:02.961Z')
        }
    ],
    work_staffs: [
        {
            alias_name: 'mik3',
            work_start: '2025-04-05T10:00',
            work_end: '2025-04-05T23:00',
            seq: 2
        },
        {
            alias_name: 'ivan1',
            work_start: '2025-04-05T15:00',
            work_end: '2025-04-05T20:00',
            seq: 2
        },
        {
            alias_name: 'mik2',
            work_start: '2025-04-05T15:00',
            work_end: '2025-04-05T20:00',
            seq: 3
        }
    ]
}