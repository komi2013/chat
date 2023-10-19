// export let publicVal = 0;

export function get_formated_time() {
  // console.log('publicFn called');

  const formated = (
      _fmt = 'YYYY/MM/DD hh:mm:ss.iii',
      _dt  = new Date(),
  ) => [
          [ 'YYYY', _dt.getFullYear()  ],
          [ 'MM',   _dt.getMonth() + 1 ], // なぜ Java と同じ仕様にしたのか？小一時間問いたい
          [ 'DD',   _dt.getDate()      ],
          [ 'hh',   _dt.getHours()     ],
          [ 'mm',   _dt.getMinutes()   ],
          [ 'ss',   _dt.getSeconds()   ],
          [ 'iii',  _dt.getMilliseconds() ],
      ].reduce(
          (s,a) => s.replace( a[0], `${a[1]}`.padStart(a[0].length,'0') ),
          _fmt
      )
  return formated()
}

// export function publicFn (
//   _fmt = 'YYYY/MM/DD hh:mm:ss.iii',
//   _dt  = new Date(),
// ) => [
//       [ 'YYYY', _dt.getFullYear()  ],
//       [ 'MM',   _dt.getMonth() + 1 ], // なぜ Java と同じ仕様にしたのか？小一時間問いたい
//       [ 'DD',   _dt.getDate()      ],
//       [ 'hh',   _dt.getHours()     ],
//       [ 'mm',   _dt.getMinutes()   ],
//       [ 'ss',   _dt.getSeconds()   ],
//       // [ 'iii',  _dt.getMilliseconds() ],
//     ].reduce(
//       (s,a) => s.replace( a[0], `${a[1]}`.padStart(a[0].length,'0') ),
//       _fmt
//     )



