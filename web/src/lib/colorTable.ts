// Color palette ported from the original VB6 m_Color.bas
// 12 base colors that get interpolated for smooth gradients.

const BASE_COLORS_RGB: [number, number, number][] = [
  [0, 0, 128],     // dark blue
  [0, 0, 255],     // blue
  [0, 128, 255],   // light blue
  [0, 255, 255],   // cyan
  [0, 255, 128],   // cyan-green
  [0, 255, 0],     // green
  [128, 255, 0],   // yellow-green
  [255, 255, 0],   // yellow
  [255, 128, 0],   // orange
  [255, 0, 0],     // red
  [255, 0, 128],   // pink
  [128, 0, 255],   // purple
];

export function buildColorTable(
  maxIterations: number,
  colorChange: number = 20,
  colorOffset: number = 3
): Uint32Array {
  const table = new Uint32Array(maxIterations + 1);
  const numColors = BASE_COLORS_RGB.length;

  let idx = 0;
  let colorIdx = colorOffset % numColors;

  while (idx < maxIterations) {
    const c1 = BASE_COLORS_RGB[colorIdx % numColors];
    const c2 = BASE_COLORS_RGB[(colorIdx + 1) % numColors];

    for (let step = 0; step < colorChange && idx < maxIterations; step++) {
      const t = step / colorChange;
      const r = Math.round(c1[0] * (1 - t) + c2[0] * t);
      const g = Math.round(c1[1] * (1 - t) + c2[1] * t);
      const b = Math.round(c1[2] * (1 - t) + c2[2] * t);
      // ABGR format for ImageData (little-endian Uint32)
      table[idx] = 0xff000000 | (b << 16) | (g << 8) | r;
      idx++;
    }
    colorIdx++;
  }

  // Last entry (max iterations = inside set) is black
  table[maxIterations] = 0xff000000;

  return table;
}
