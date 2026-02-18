<template>
  <div class="flex h-screen min-h-0 flex-col bg-slate-50 text-slate-900">
    <div class="shrink-0 px-4 pb-2 pt-4">
      <slot name="top" />
    </div>
    <main class="min-h-0 flex-1 overflow-hidden">
      <slot />
    </main>
  </div>
</template>
