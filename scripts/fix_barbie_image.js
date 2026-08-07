/**
 * Quick fix: sets image_url for "Barbie Doll" (the one product missing an image).
 * Usage: node fix_barbie_image.js --token YOUR_ADMIN_TOKEN --images "C:\Users\Administrator\Desktop\Blinkit\assets"
 */

const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://ecommerce-backend-dd4u.onrender.com/api/v1';
const PRODUCT_NAME = 'Barbie Doll';
const IMAGE_FILENAME = 'toys1.png';

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {};
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--token') out.token = args[++i];
    if (args[i] === '--images') out.imagesDir = args[++i];
  }
  return out;
}

function findFileRecursive(rootDir, targetName) {
  const stack = [rootDir];
  while (stack.length) {
    const dir = stack.pop();
    let entries;
    try {
      entries = fs.readdirSync(dir, { withFileTypes: true });
    } catch (e) {
      continue;
    }
    for (const entry of entries) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        stack.push(full);
      } else if (entry.name.toLowerCase() === targetName.toLowerCase()) {
        return full;
      }
    }
  }
  return null;
}

async function main() {
  const { token, imagesDir } = parseArgs();
  if (!token || !imagesDir) {
    console.error('Usage: node fix_barbie_image.js --token YOUR_ADMIN_TOKEN --images "path\\to\\Blinkit\\assets"');
    process.exit(1);
  }

  const headers = { Authorization: `Bearer ${token}` };

  console.log('Fetching products to find Barbie Doll...');
  let product = null;
  let page = 1;
  while (!product) {
    const res = await fetch(`${BASE_URL}/products/?limit=100&page=${page}`, { headers });
    const data = await res.json();
    const items = data.products || [];
    product = items.find((p) => p.name === PRODUCT_NAME);
    if (product || page >= (data.total_pages || 1)) break;
    page++;
  }

  if (!product) {
    console.error(`Product "${PRODUCT_NAME}" not found in backend.`);
    process.exit(1);
  }
  console.log(`Found: ${product.name} (id ${product.id}, category_id ${product.category_id})`);

  const localPath = findFileRecursive(imagesDir, IMAGE_FILENAME);
  if (!localPath) {
    console.error(`Local image "${IMAGE_FILENAME}" not found under ${imagesDir}`);
    process.exit(1);
  }
  console.log(`Using image: ${localPath}`);

  const fileBuffer = fs.readFileSync(localPath);
  const ext = path.extname(localPath).toLowerCase();
  const mime = ext === '.jpg' || ext === '.jpeg' ? 'image/jpeg' : 'image/png';
  const blob = new Blob([fileBuffer], { type: mime });
  const form = new FormData();
  form.append('image', blob, path.basename(localPath));

  const uploadRes = await fetch(`${BASE_URL}/upload`, { method: 'POST', headers, body: form });
  const uploadData = await uploadRes.json();
  if (!uploadRes.ok) {
    console.error('Upload failed:', uploadData);
    process.exit(1);
  }
  const imageUrl = uploadData.url || uploadData.image_url;
  console.log(`Uploaded, image_url: ${imageUrl}`);

  const body = {
    name: product.name,
    description: product.description || product.name,
    price: product.price,
    image_url: imageUrl,
    category_id: product.category_id,
    stock: product.stock ?? 50,
  };

  let updRes = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
    method: 'PUT',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (updRes.status === 404 || updRes.status === 405) {
    updRes = await fetch(`${BASE_URL}/admin/products/${product.id}`, {
      method: 'PATCH',
      headers: { ...headers, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }
  const updData = await updRes.json().catch(() => ({}));
  if (!updRes.ok) {
    console.error('Update failed:', updData);
    process.exit(1);
  }
  console.log('Done! Barbie Doll image updated successfully.');
}

main();
