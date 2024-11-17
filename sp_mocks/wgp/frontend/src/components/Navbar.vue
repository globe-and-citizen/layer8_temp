<script setup>
import {ref} from "vue"

const counter = ref(0)

async function persistenceCheckHandler (){
  let res = await layer8.persistenceCheck(">ARGUMENT PASSED IN<")
  counter.value = res
}

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL
async function pingBackend() {
  try {
    let response = await layer8.fetch(BACKEND_URL, {
      method: "POST",
      headers: {
        "Content-Type": "Application/Json"
      },
      body: JSON.stringify({
        email: "registerEmail.value",
        password: "registerPassword.value"
      })
    });
    let rawHeaderObject = {}
    response.headers.forEach((val, key) => {
      rawHeaderObject[key] = val
    })
    console.log("Ping 8000 - await response.text(): ", await response.text())
    console.log("Ping 8000 - response.status: ", response.status)
    console.log("Ping 8000 - rawHeaderObject: ", rawHeaderObject)

  } catch (error) {
    console.log("Ping to 8000 failed from navbar: ", error);
  }
};

</script>

<template>
  <div class="navbar bg-base-100">
    <div class="flex-1">
      <a class="btn btn-ghost text-xl">SP MOCK</a>
    </div>
    <div class="flex-none">
      <ul class="menu menu-horizontal px-1 bg-base-100">
        <li>
          <RouterLink to="/">Home</RouterLink>
        </li>
        <li>
          <RouterLink to="/stress-test">Stress test</RouterLink>
        </li>
        <!-- <li> <button @click="pingBackend">Ping 8000</button>  </li>-->
        <li class="inline-block">
          <button @click="persistenceCheckHandler">Check WASM Persistence
            <span v-if="counter != 0">{{ counter }}</span></button>
        </li>
      </ul>
    </div>
  </div>
</template>
<style>
a, li {
  color: #000 !important;
}
</style>