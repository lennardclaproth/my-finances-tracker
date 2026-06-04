<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Sparkline from './Sparkline.svelte';

  import { sparklineTones } from './sparkline.types';

  const rising = [4, 6, 5, 8, 7, 10, 12, 11, 14];
  const falling = [14, 12, 13, 9, 10, 7, 6, 8, 4];

  const { Story } = defineMeta({
    title: 'Atoms/Sparkline',
    component: Sparkline,
    tags: ['autodocs'],
    argTypes: {
      data: { control: 'object' },
      width: { control: 'number' },
      height: { control: 'number' },
      strokeWidth: { control: 'number' },
      tone: { control: 'select', options: sparklineTones },
      fill: { control: 'boolean' },
      ariaLabel: { control: 'text' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    data: rising,
    width: 120,
    height: 36,
    strokeWidth: 2,
    tone: 'auto',
    fill: false,
    ariaLabel: 'Balance trend'
  }}
/>

<Story name="Trends" asChild>
  <div class="flex items-center gap-6 text-slate-700">
    <Sparkline data={rising} ariaLabel="Rising balance" />
    <Sparkline data={falling} ariaLabel="Falling balance" />
  </div>
</Story>

<Story name="With Fill" asChild>
  <div class="flex items-center gap-6">
    <Sparkline data={rising} fill ariaLabel="Rising balance" />
    <Sparkline data={falling} fill ariaLabel="Falling balance" />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-6 text-slate-700">
    <Sparkline data={rising} width={80} height={24} />
    <Sparkline data={rising} width={160} height={48} strokeWidth={2.5} />
  </div>
</Story>
