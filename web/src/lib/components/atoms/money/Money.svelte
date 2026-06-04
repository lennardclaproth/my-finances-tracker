<script lang="ts">
  import type {
    MoneySignDisplay,
    MoneySize,
    MoneyWeight
  } from './money.types';

  import {
    moneyBaseClasses,
    moneySignClasses,
    moneySizeClasses,
    moneyWeightClasses
  } from './money.variants';

  type Props = {
    /** Monetary value in major currency units (e.g. 12.5 renders as €12.50). */
    amount: number;
    /** ISO 4217 currency code used for formatting. */
    currency?: string;
    /** BCP 47 locale controlling number/currency formatting. */
    locale?: string;
    size?: MoneySize;
    weight?: MoneyWeight;
    signDisplay?: MoneySignDisplay;
    /** Colour the value by sign (positive / negative / zero). */
    colored?: boolean;
    class?: string;
  };

  let {
    amount,
    currency = 'EUR',
    locale = 'en-US',
    size = 'md',
    weight = 'medium',
    signDisplay = 'auto',
    colored = false,
    class: className = ''
  }: Props = $props();

  const formatted = $derived(
    new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      signDisplay
    }).format(amount)
  );

  const sign = $derived(amount > 0 ? 'positive' : amount < 0 ? 'negative' : 'zero');

  const classes = $derived([
    moneyBaseClasses,
    moneySizeClasses[size],
    moneyWeightClasses[weight],
    colored ? moneySignClasses[sign] : '',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<span class={classes}>{formatted}</span>
