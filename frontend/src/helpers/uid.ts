export function generateSecureUid(length: number): string {
  const chars = "0123456789abcdefghijklmnopqrstuvwxyz";
  const array = new Uint8Array(length);

  // Use Web Crypto API for cryptographically secure randomness
  (window.crypto || self.crypto).getRandomValues(array);

  let id = "";

  for (let i = 0; i < array.length; i++) {
    id += chars[array[i] % chars.length];
  }

  return id;
}
