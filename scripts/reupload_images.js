/**
 * Re-uploads local product images (now that /upload goes to Cloudinary
 * instead of ephemeral local disk) and updates each existing product's
 * image_url to the new permanent URL.
 *
 * Only touches products whose name matches an entry in blinkit_products.json
 * that has a real local image file (i.e. the original 244 "batch 1"
 * products). Batch 2 products never had local images, so there's nothing
 * to re-upload for them - they'll just keep showing the placeholder icon
 * until real photos are added.
 *
 * USAGE:
 *   node reupload_images.js --token YOUR_ADMIN_TOKEN --images "C:\Users\Administrator\Desktop\Blinkit\assets"
 */

const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {};
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--token') out.token = args[++i];
    if (args[i] === '--images') out.imagesDir = args[++i];
  }
  return out;
}

async function fetchAllProducts(headers) {
  const all = [];
  let page = 1;
  while (true) {
    const res = await fetch(`${BASE_URL}/products?limit=100&page=${page}`);
    const data = await res.json();
    all.push(...(data.products || []));
    if (page >= (data.total_pages || 1)) break;
    page++;
  }
  return all;
}

async function main() {
  const { token, imagesDir } = parseArgs();
  if (!token || !imagesDir) {
    console.error('Usage: node reupload_images.js --token YOUR_ADMIN_TOKEN --images "path\\to\\Blinkit\\assets"');
    process.exit(1);
  }

  const catalog1 = JSON.parse(fs.readFileSync(path.join(__dirname, 'blinkit_products.json'), 'utf-8'));
  let catalog2 = [];
  const batch2Path = path.join(__dirname, 'blinkit_products_batch2.json');
  if (fs.existsSync(batch2Path)) {
    catalog2 = JSON.parse(fs.readFileSync(batch2Path, 'utf-8'));
  }
  const imageByName = {};
  for (const p of [...catalog1, ...catalog2]) imageByName[p.name.toLowerCase()] = p.image;

  const headers = { Authorization: `Bearer ${token}` };

  console.log('Fetching existing products from backend...');
  const products = await fetchAllProducts(headers);
  console.log(`Found ${products.length} products in backend.`);

  let updated = 0, skippedNoLocalImage = 0, failed = 0;

  for (const product of products) {
    if (product.image_url && product.image_url.startsWith('http')) {
      continue; // already has a permanent Cloudinary URL from a previous run
    }
    const catalogImage = imageByName[product.name.toLowerCase()];
    if (!catalogImage) {
      skippedNoLocalImage++;
      continue;
    }
    const relativePath = catalogImage.startsWith('assets/') ? catalogImage.slice('assets/'.length) : catalogImage;
    const localImagePath = path.join(imagesDir, relativePath);
    if (!fs.existsSync(localImagePath)) {
      skippedNoLocalImage++;
      continue;
    }

    try {
      const fileBuffer = fs.readFileSync(localImagePath);
      const ext = path.extname(localImagePath).toLowerCase();
      const mime = ext === '.jpg' || ext === '.jpeg' ? 'image/jpeg' : ext === '.webp' ? 'image/webp' : 'image/png';
      const blob = new Blob([fileBuffer], { type: mime });
      const form = new FormData();
      form.append('image', blob, path.basename(localImagePath));

      const uploadRes = await fetch(`${BASE_URL}/upload`, {
        method: 'POST',
        headers,
        body: form,
      });
      const uploadData = await uploadRes.json();
      if (!uploadRes.ok || !uploadData.image_url) {
        console.error(`  FAIL ${product.name}: upload failed - ${uploadData.error || uploadRes.status}`);
        failed++;
        continue;
      }

      const putRes = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
        method: 'PUT',
        headers: { ...headers, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: product.name,
          description: product.description,
          price: product.price,
          image_url: uploadData.image_url,
          category_id: product.category_id,
        }),
      });
      const putData = await putRes.json();
      if (putRes.ok) {
        updated++;
        console.log(`  OK  ${product.name} -> ${uploadData.image_url}`);
      } else {
        failed++;
        console.error(`  FAIL ${product.name}: update failed - ${putData.error || putRes.status}`);
      }
    } catch (e) {
      failed++;
      console.error(`  FAIL ${product.name}: ${e.message}`);
    }
  }

  console.log(`\nDone. Updated: ${updated}, Failed: ${failed}, No local image found: ${skippedNoLocalImage}`);
}

main();