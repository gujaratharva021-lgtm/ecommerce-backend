const BASE = "http://192.168.1.12:8081/api/v1";
const UPLOAD_BASE = "http://192.168.1.12:8081";

function parseArgs() {
  const args = process.argv.slice(2);
  const out = {};
  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--token") out.token = args[++i];
  }
  return out;
}

const IMAGE_MAP = {
  "Cola 750ml": "cola_750ml.png",
  "Orange Juice 1L": "orange_juice_1l.png",
  "Marie Gold Biscuits": "marie_gold_biscuits.jpg",
  "Cream Biscuits": "cream_biscuits.jpg",
  "Dark Chocolate Bar": "dark_chocolate_bar.jpg",
  "Milk Chocolate Bar": "milk_chocolate_bar.png",
  "Fresh Apples 1kg": "fresh_apples_1kg.png",
  "Bananas 1 dozen": "bananas_1_dozen.png",
  "Vanilla Ice Cream 500ml": "vanilla_icecream_500ml.jpg",
  "Chocolate Ice Cream 500ml": "chocolate_icecream_500ml.png",
  "Tomato Ketchup 500g": "tomato_ketchup_500g.jpg",
  "Chilli Ketchup 350g": "chilli_ketchup_350g.jpg",
  "Aloo Bhujia 200g": "aloo_bhujia_200g.jpg",
  "Mixture 200g": "mixture_200g.jpg",
  "Hand Sanitizer 100ml": "hand_sanitizer_100ml.png",
  "Face Wash 100ml": "face_wash_100ml.jpg",
  "Mango Pickle 400g": "mango_pickle_400g.png",
  "Mixed Veg Pickle 400g": "mixed_veg_pickle_400g.jpg",
  "Agarbatti Pack": "agarbatti_pack.png",
  "Diya Set (Pack of 12)": "diya_set.jpg",
  "Anti-Dandruff Shampoo 200ml": "antidandruff_shampoo_200ml.jpg",
  "Herbal Shampoo 340ml": "herbal_shampoo_340ml.jpg",
  "Potato Chips 100g": "potato_chips_100g.jpg",
  "Peanut Snacks 150g": "peanut_snacks_150g.jpg",
  "Sandalwood Soap": "sandalwood_soap.jpg",
  "Neem Soap": "neem_soap.png",
  "Building Blocks Set": "building_blocks_set.png",
  "Remote Control Car": "remote_control_car.jpg",
  "Chocolate Wafers 100g": "chocolate_wafers_100g.jpg",
  "Cream Wafers 100g": "cream_wafers_100g.jpg",
};

async function main() {
  const { token } = parseArgs();
  if (!token) {
    console.error("Usage: node fix_local_images.js --token YOUR_ADMIN_TOKEN");
    process.exit(1);
  }

  const headers = {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };

  console.log("Fetching all products...");
  let all = [];
  let page = 1;
  const limit = 200;
  while (page <= 100) {
    const res = await fetch(`${BASE}/products?page=${page}&limit=${limit}`);
    const data = await res.json();
    const batch = data.products || [];
    if (batch.length === 0) break;
    all = all.concat(batch);
    page++;
  }
  console.log(`Fetched ${all.length} products total.\n`);

  let ok = 0, fail = 0, skipped = 0;
  for (const p of all) {
    const filename = IMAGE_MAP[p.name];
    if (!filename) {
      skipped++;
      continue;
    }
    const imageUrl = `${UPLOAD_BASE}/uploads/${filename}`;
    const body = {
      name: p.name,
      description: p.description,
      price: p.price,
      image_url: imageUrl,
      category_id: p.category_id,
      stock: p.stock,
    };
    try {
      const res = await fetch(`${BASE}/admin/products/${p.id}`, {
        method: "PUT",
        headers,
        body: JSON.stringify(body),
      });
      if (res.ok) {
        ok++;
        console.log(`  FIXED id ${p.id}: ${p.name} -> ${imageUrl}`);
      } else {
        fail++;
        const data = await res.json().catch(() => ({}));
        console.error(`  FAIL id ${p.id} (${p.name}):`, data.error || res.status);
      }
    } catch (e) {
      fail++;
      console.error(`  FAIL id ${p.id} (${p.name}):`, e.message);
    }
  }

  console.log(`\nDone. Fixed: ${ok}, Failed: ${fail}, Skipped (not in list): ${skipped}`);
}

main().catch((e) => {
  console.error("Fatal error:", e);
  process.exit(1);
});
