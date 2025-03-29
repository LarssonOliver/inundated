export function shouldTextBeDarkFromBgColor(bgColor: string): boolean {
  let hexColor = bgColor.replace("#", "");

  if (hexColor.length === 3) {
    hexColor = hexColor
      .split("")
      .map((char) => char + char)
      .join("");
  }

  if (hexColor.length !== 6) {
    return false;
  }

  const r = parseInt(hexColor.slice(0, 2), 16);
  const g = parseInt(hexColor.slice(2, 4), 16);
  const b = parseInt(hexColor.slice(4, 6), 16);

  return luminance(r, g, b) > 0.179;
}

function luminance(r: number, g: number, b: number): number {
  const normalizedComponents = [r, g, b].map((c) => c / 255);

  const [sR, sG, sB] = normalizedComponents.map((c) =>
    c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4),
  );

  return 0.2126 * sR + 0.7152 * sG + 0.0722 * sB;
}

export function stringToHexColor(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  let color = "#";
  for (let i = 0; i < 3; i++) {
    const value = (hash >> (i * 8)) & 0xff;
    color += ("00" + value.toString(16)).slice(-2);
  }
  return color;
}
