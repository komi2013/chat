export async function receptionOrder(pushData) {
  const pushID = pushData[0];
  const fd = new FormData();
  fd.append('pushID', pushID);
  const request = new Request('/PushResponse/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)

  const data = pushData[4]
  if (Array.isArray(data)) {
    // case: [ { receptionOrderID: ..., seatName: ... }, ... ]
    for (const order of data) {
      await upsertIDB(order, 'receptionOrder', 'receptionOrderID', order.receptionOrderID)
    }
  } else if (typeof data === 'string') {
    // case: "seatName"
    const seatName = data
    const orders = await getAllIDBs('receptionOrder')
    const orderSeats = orders.filter(order => order.seatName === seatName)

    for (const order of orderSeats) {
      await deleteIDB('receptionOrder', 'receptionOrderID', order.receptionOrderID)
    }
  }


  // const orders = pushData[4]
  // console.log('orders', orders)
  // if (Array.isArray(orders)) {
  //   for (const order of orders) {
  //     await upsertIDB(order, 'receptionOrder', 'receptionOrderID', order.receptionOrderID)
  //   }
  // }

  // const orders = await getAllIDBs('receptionOrder')
  // const orderSeats = orders.filter(order => order.seatName === seatName)

  // if (Array.isArray(orderSeats)) {
  //   for (const order of orderSeats) {
  //     await deleteIDB('receptionOrder', 'receptionOrderID', order.receptionOrderID)
  //   }
  // }


  // if (pushData[3]) {
  //   const receptionOrder = {
  //     receptionOrderID: pushData[2] + '_' + pushData[3] + '_' + pushData[6],
  //     tableName: pushData[2],
  //     menuID: pushData[3],
  //     itemDetailIDs: pushData[4],
  //     price: pushData[5]
  //   }
  //   upsertIDB(receptionOrder, 'receptionOrder', 'receptionOrderID', receptionOrder.receptionOrderID)
  // } else {

  //   const allData = await getAllIDBs('receptionOrder');
  //   const targetTableName = pushData[2];
  //   const deletePromises = allData
  //     .filter(item => item.tableName === targetTableName) // 条件に一致するデータをフィルタ
  //     .map(item => deleteIDB('receptionOrder', 'receptionOrderID', item.receptionOrderID)); 

  //   const deleteResults = await Promise.all(deletePromises);

  //   console.log(deleteResults);
  //   console.log('Matching data deleted successfully.');

  //   // deleteIDB('receptionOrder', 'receptionOrderID', receptionOrder.receptionOrderID);
  // }
}