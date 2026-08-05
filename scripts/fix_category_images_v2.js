const fs = require("fs");
const path = require("path");

const ASSETS_DIR = "C:\\Users\\Administrator\\Desktop\\Blinkit\\assets\\categories";
const UPLOADS_DIR = "C:\\Users\\Administrator\\Desktop\\ecommerce-backend\\uploads";
const BASE = "http://192.168.1.12:8081/api/v1";
const UPLOAD_BASE = "http://192.168.1.12:8081";

const CATEGORY_IMAGES = {
  12: { name: "Bakery", file: "bakery.png" },
  10: { name: "Beverages", file: "beverages.png" },
  13: { name: "Biscuits", file: "biscuits.png" },
  9:  { name: "Chocolate", file: "chocolate.png" },
  8:  { name: "Fruits", file: "fruits.png" },
  11: { name: "Ice Creams", file: "ice_creams.png" },
  16: { name: "Ketchup", file: "ketchup.png" },
  14: { name: "Namkeen", file: "namkeen.png" },
  19: { name: "Personal Care", file: "personal_care.png" },
  20: { name: "Pickle", file: "pickle.png" },
  21: { name: "Puja Items", file: "puja_items.png" },
  17: { name: "Shampoo", file: "shampoo.png" },
  18: { name: "Soap", file: "soap.png" },
  22: { name: "Toys", file: "toys.png" },
  15: { name: "Wafers", file: "wafers.png" },
};

async function main() {
  const idx = process.argv.indexOf("--token");
  const token = idx >= 0 ? process.argv[idx + 1] : null;
  if (!token) {
    console.error("Usage: node fix_category_images_v2.js --token YOUR_ADMIN_TOKEN");
    process.exit(1);
  }
  const headers = { Authorization: `Bearer ${token}`, "Content-Type": "application/json" };

  fs.mkdirSync(UPLOADS_DIR, { recursive: true });

  let ok = 0, fail = 0;
  for (const [catId, info] of Object.entries(CATEGORY_IMAGES)) {
    const srcPath = path.join(ASSETS_DIR, info.file);
    if (!fs.existsSync(srcPath)) {
      console.error(`  MISSING SOURCE: ${info.file}`);
      fail++;
      continue;
    }
    const ext = path.extname(srcPath);
    const destFilename = `category_${catId}_v2${ext}`;
    const destPath = path.join(UPLOADS_DIR, destFilename);
    fs.copyFileSync(srcPath, destPath);

    const imageUrl = `${UPLOAD_BASE}/uploads/${destFilename}`;
    try {
      const res = await fetch(`${BASE}/admin/categories/${catId}`, {
        method: "PUT",
        headers,
        body: JSON.stringify({ name: info.name, image_url: imageUrl }),
      });
      if (res.ok) {
        ok++;
        console.log(`  UPDATED category ${catId} (${info.name}) -> ${imageUrl}`);
      } else {
        fail++;
        const data = await res.json().catch(() => ({}));
        console.error(`  FAIL category ${catId} (${info.name}):`, data.error || res.status);
      }
    } catch (e) {
      fail++;
      console.error(`  FAIL category ${catId} (${info.name}):`, e.message);
    }
  }

  console.log(`\nDone. Updated: ${ok}, Failed: ${fail}`);
}

main().catch((e) => {
  console.error("Fatal error:", e);
  process.exit(1);
});
