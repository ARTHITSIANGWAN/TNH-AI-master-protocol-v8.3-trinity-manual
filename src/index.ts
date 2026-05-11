export default {
  async fetch(request, env, ctx) {
    // 1. ทดสอบบันทึกสติลง KV
    await env.TNH_KV.put('LAST_BOOT', new Date().toISOString());

    // 2. ทดสอบดึงข้อมูลจากสมอง D1
    const { results } = await env.DB.prepare("SELECT name FROM sqlite_master LIMIT 1").all();

    // 3. เตรียมเรียกใช้ AI (สมมติใช้ GPT-5.5 รุ่นใหม่)
    // const aiResponse = await env.AI.run('@cf/openai/gpt-5.5', { prompt: "Hello Trinity!" });

    return new Response(
      JSON.stringify({
        status: "TNH-AI-V83-ONLINE",
        message: "อาณาจักร 9THERA พร้อมรบ!",
        kv_check: await env.TNH_KV.get('LAST_BOOT'),
        db_ready: !!results
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }
};
