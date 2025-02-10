import { pushReceive } from './pushReceive.js';

export async function chunk(pushData) {
  const chunk = {
    chunkID: pushData[3] + pushData[4],
    strChunk: pushData[2],
    chunkPass: pushData[3],
    chunkIndex: pushData[4],
    chunkLength: pushData[5]
  };

  const chunks = await getIDBs('chunk', 'chunkPassIndex', chunk.chunkPass, 1000)
  chunks.push(chunk);
  if (chunks.length == chunk.chunkLength) {
    const sortedData = chunks.sort((a, b) => a.chunkIndex - b.chunkIndex);
    const combinedString = sortedData.map(item => item.strChunk).join('');
    let data = JSON.parse(combinedString);
    data.unshift("");
    pushReceive(JSON.stringify(data));
  } else {
    upsertIDB(chunk, 'chunk', 'chunkID', chunk.chunkID)
      .catch((error) => {
        console.error(error);
      });
  }
}

