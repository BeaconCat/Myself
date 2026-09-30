/**
 * 只读 zip 的中央目录，列出条目名（不解压、不读全文件）：
 * 从文件末尾找 EOCD（含 ZIP64），再按目录偏移读中央目录。用于上传前的快速校验，比如「是不是本站备份包」。
 * 不是 zip 返回空数组；目录损坏等解析不了的情况返回 null，由调用方交给服务端判断。
 */

const EOCD = 0x06054b50;
const EOCD64 = 0x06064b50;
const EOCD64_LOC = 0x07064b50;
const CEN = 0x02014b50;
/** 中央目录读取上限：备份包里素材再多，目录也远小于这个值 */
const MAX_DIR = 64 << 20;

async function read(file: Blob, start: number, end: number): Promise<DataView> {
  return new DataView(await file.slice(start, end).arrayBuffer());
}

function u64(v: DataView, at: number): number {
  return v.getUint32(at, true) + v.getUint32(at + 4, true) * 2 ** 32;
}

export async function zipEntries(file: Blob): Promise<string[] | null> {
  try {
    // EOCD 固定 22 字节 + 最多 64 KB 注释
    const tailStart = Math.max(0, file.size - (22 + 0xffff));
    const tail = await read(file, tailStart, file.size);
    let at = -1;
    for (let i = tail.byteLength - 22; i >= 0; i--) {
      if (tail.getUint32(i, true) === EOCD) {
        at = i;
        break;
      }
    }
    // 找不到目录结尾：根本不是 zip
    if (at < 0) return [];
    let count = tail.getUint16(at + 10, true);
    let size = tail.getUint32(at + 12, true);
    let offset = tail.getUint32(at + 16, true);
    // ZIP64：字段溢出时读 ZIP64 EOCD 定位器（紧挨在 EOCD 前 20 字节）
    if (count === 0xffff || size === 0xffffffff || offset === 0xffffffff) {
      const locAt = tailStart + at - 20;
      if (locAt < 0) return null;
      const loc = await read(file, locAt, locAt + 20);
      if (loc.getUint32(0, true) !== EOCD64_LOC) return null;
      const e64At = u64(loc, 8);
      const e64 = await read(file, e64At, e64At + 56);
      if (e64.getUint32(0, true) !== EOCD64) return null;
      count = u64(e64, 32);
      size = u64(e64, 40);
      offset = u64(e64, 48);
    }
    if (size > MAX_DIR || offset + size > file.size) return null;
    const dir = await read(file, offset, offset + size);
    const dec = new TextDecoder();
    const names: string[] = [];
    let p = 0;
    for (let n = 0; n < count && p + 46 <= dir.byteLength; n++) {
      if (dir.getUint32(p, true) !== CEN) return null;
      const nameLen = dir.getUint16(p + 28, true);
      const extraLen = dir.getUint16(p + 30, true);
      const commentLen = dir.getUint16(p + 32, true);
      names.push(dec.decode(new Uint8Array(dir.buffer, dir.byteOffset + p + 46, nameLen)));
      p += 46 + nameLen + extraLen + commentLen;
    }
    return names;
  } catch {
    return null;
  }
}

/** 本站备份包：根目录下有 data/myself.db（与服务端 looksLikeBackup 同一判定）；无法判断时返回 null */
export async function looksLikeBackup(file: Blob): Promise<boolean | null> {
  const names = await zipEntries(file);
  if (!names) return null;
  return names.some((n) => n.replace(/\\/g, '/').replace(/^\.\//, '') === 'data/myself.db');
}
