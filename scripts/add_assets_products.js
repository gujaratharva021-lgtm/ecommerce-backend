const fs = require("fs");
const path = require("path");

const ASSETS_DIR = "C:\\Users\\Administrator\\Desktop\\Blinkit\\assets\\images";
const UPLOADS_DIR = "C:\\Users\\Administrator\\Desktop\\ecommerce-backend\\uploads";
const BASE = "http://192.168.1.12:8081/api/v1";
const UPLOAD_BASE = "http://192.168.1.12:8081";

const FOLDER_CATEGORY = {
  "Bakery": 12,
  "Beverages": 10,
  "Biscuits": 13,
  "Chocolate": 9,
  "cloths items": 23,
  "Fruits": 8,
  "Ice Creams": 11,
  "ketchup": 16,
  "Namkeen": 14,
  "personal use items": 19,
  "pickel": 20,
  "puja items": 21,
  "shampoo": 17,
  "Soap": 18,
  "toys items": 22,
  "Wafers": 15,
};

const PRICE_RANGE = {
  12: [40, 250], 10: [20, 150], 13: [15, 120], 9: [30, 300],
  23: [199, 1999], 8: [40, 300], 11: [80, 400], 16: [40, 200],
  14: [20, 150], 19: [50, 400], 20: [60, 250], 21: [20, 300],
  17: [80, 400], 18: [20, 120], 22: [150, 2000], 15: [15, 100],
};

function sanitizeFilename(f) {
  return f.replace(/[^a-zA-Z0-9.\-_]/g, "_");
}

function cleanName(filename) {
  let name = filename.replace(/\.[^/.]+$/, "");
  name = name.replace(/\s*\(\d+\)\s*$/, "");
  name = name.replace(/[_-]+/g, " ").trim();
  name = name.replace(/\s+/g, " ");
  return name;
}

async function main() {
  const idx = process.argv.indexOf("--token");
  const token = idx >= 0 ? process.argv[idx + 1] : null;
  if (!token) {
    console.error("Usage: node add_assets_products.js --token YOUR_ADMIN_TOKEN");
    process.exit(1);
  }
  const headers = { Authorization: `Bearer ${token}`, "Content-Type": "application/json" };

  fs.mkdirSync(UPLOADS_DIR, { recursive: true });

  let created = 0, failed = 0, skippedDup = 0;

  for (const [folder, categoryId] of Object.entries(FOLDER_CATEGORY)) {
    const srcFolder = path.join(ASSETS_DIR, folder);
    if (!fs.existsSync(srcFolder)) {
      console.log(`Skipping missing folder: ${folder}`);
      continue;
    }
    const files = fs.readdirSync(srcFolder).filter(f => /\.(jpg|jpeg|png|webp)$/i.test(f));
    const seenNames = new Set();
    console.log(`\n=== ${folder} (${files.length} files) ===`);
    for (const file of files) {
      const productName = cleanName(file);
      const key = productName.toLowerCase();
      if (seenNames.has(key)) { skippedDup++; continue; }
      seenNames.add(key);

      const destFilename = `${folder.replace(/\s+/g, "_")}_${sanitizeFilename(file)}`;
      const destPath = path.join(UPLOADS_DIR, destFilename);
      try {
        fs.copyFileSync(path.join(srcFolder, file), destPath);
      } catch (e) {
        console.error(`  COPY FAIL ${file}: ${e.message}`);
        failed++;
        continue;
      }

      const imageUrl = `${UPLOAD_BASE}/uploads/${encodeURIComponent(destFilename)}`;
      const [min, max] = PRICE_RANGE[categoryId] || [50, 300];
      const price = Math.floor(Math.random() * (max - min + 1)) + min;

      const body = {
        name: productName,
        description: `${productName} - quality product, freshly stocked.`,
        price: price,
        image_url: imageUrl,
        category_id: categoryId,
        stock: 50,
      };

      try {
        const res = await fetch(`${BASE}/admin/products`, {
          method: "POST",
          headers,
          body: JSON.stringify(body),
        });
        if (res.ok) {
          created++;
          console.log(`  CREATED: ${productName} (Rs.${price})`);
        } else {
          failed++;
          const data = await res.json().catch(() => ({}));
          console.error(`  FAIL: ${productName} ->`, data.error || res.status);
        }
      } catch (e) {
        failed++;
        console.error(`  FAIL: ${productName} ->`, e.message);
      }
    }
  }

  console.log(`\nDone. Created: ${created}, Failed: ${failed}, Skipped duplicates: ${skippedDup}`);
}

main().catch((e) => {
  console.error("Fatal error:", e);
  process.exit(1);
});
