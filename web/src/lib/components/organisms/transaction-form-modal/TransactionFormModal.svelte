<script lang="ts">
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import type { CashflowTransactionFormValue } from './transaction-form-modal.types';

	type Props = {
		open?: boolean;
		title?: string;
		submitting?: boolean;
		error?: string | null;
		onSubmit?: (value: CashflowTransactionFormValue) => void;
		onClose?: () => void;
	};

	let {
		open = $bindable(false),
		title = 'New transaction',
		submitting = false,
		error = null,
		onSubmit,
		onClose
	}: Props = $props();

	let date = $state(todayISO());
	let amount = $state('');
	let type = $state('expense');
	let description = $state('');
	let note = $state('');
	let tag = $state('');
	let attempted = $state(false);

	// Reset the form each time the modal opens.
	$effect(() => {
		if (open) {
			date = todayISO();
			amount = '';
			type = 'expense';
			description = '';
			note = '';
			tag = '';
			attempted = false;
		}
	});

	const amountError = $derived(
		attempted && amount.trim() === '' ? 'Amount is required' : undefined
	);
	const descriptionError = $derived(
		attempted && description.trim() === '' ? 'Description is required' : undefined
	);

	const typeOptions = [
		{ value: 'expense', label: 'Expense' },
		{ value: 'income', label: 'Income' }
	];

	function submit() {
		attempted = true;
		if (amount.trim() === '' || description.trim() === '') return;
		onSubmit?.({
			date,
			amount: amount.trim(),
			type: type as CashflowTransactionFormValue['type'],
			description: description.trim(),
			note: note.trim(),
			tag: tag.trim()
		});
	}
</script>

<Dialog bind:open {title} size="md" {onClose}>
	<div class="space-y-3">
		{#if error}
			<div class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
				{error}
			</div>
		{/if}

		<div class="grid grid-cols-2 gap-3">
			<FormField label="Date" id="tx-date">
				<DatePicker value={date} onChange={(v) => (date = v ?? date)} class="w-full" />
			</FormField>

			<FormField label="Type" id="tx-type">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={type}
						options={typeOptions}
						ariaLabel="Transaction type"
					/>
				{/snippet}
			</FormField>
		</div>

		<FormField label="Amount" id="tx-amount" error={amountError}>
			{#snippet children(ctx)}
				<CurrencyInput
					id={ctx.id}
					bind:value={amount}
					intent={ctx.invalid ? 'error' : 'default'}
					ariaDescribedby={ctx.describedby}
				/>
			{/snippet}
		</FormField>

		<FormField label="Description" id="tx-description" error={descriptionError}>
			{#snippet children(ctx)}
				<Input
					id={ctx.id}
					bind:value={description}
					placeholder="e.g. Albert Heijn"
					intent={ctx.invalid ? 'error' : 'default'}
					ariaDescribedby={ctx.describedby}
				/>
			{/snippet}
		</FormField>

		<div class="grid grid-cols-2 gap-3">
			<FormField label="Tag" id="tx-tag" hint="Optional">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={tag} placeholder="e.g. groceries" />
				{/snippet}
			</FormField>

			<FormField label="Note" id="tx-note" hint="Optional">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={note} />
				{/snippet}
			</FormField>
		</div>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Cancel</Button>
		<Button onclick={submit} loading={submitting}>Save</Button>
	{/snippet}
</Dialog>
