export default {
  async fetch(request, env, ctx) {
    // ใช้ TNH_KV แทน KV เฉยๆ เพื่อความกริบ
    await env.TNH_KV.put('STATUS', 'MISSION_START_11:55');
    
    const value = await env.TNH_KV.get('STATUS');
    const allKeys = await env.TNH_KV.list();

    return new Response(
      JSON.stringify({
        project: "TNH-AI-V8.3",
        status: value,
        allKeys: allKeys,
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  } 
}

