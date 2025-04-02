<template>
    <div
        class="flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8"
    >
        <UCard class="w-full max-w-md bg-gray-900">
            <div class="flex flex-col items-center">
                <h1
                    class="text-center text-2xl font-bold tracking-tight text-white"
                >
                    Sign in to your account
                </h1>
                <p class="mt-2 text-center text-sm text-gray-200">
                    Or
                    <NuxtLink
                        to="/register"
                        class="font-medium text-primary hover:text-primary-dark underline"
                    >
                        create a new account
                    </NuxtLink>
                </p>
            </div>

            <div class="mt-8">
                <form
                    @submit.prevent="handleSubmit"
                    class="space-y-6"
                    method="POST"
                >
                    <UFormField label="Username" name="username">
                        <UInput
                            v-model="username"
                            name="username"
                            autocomplete="username"
                            required
                            placeholder="mario.rossi"
                            class="w-full"
                        />
                    </UFormField>

                    <UFormField label="Password" name="password">
                        <UInput
                            v-model="password"
                            type="password"
                            name="password"
                            autocomplete="current-password"
                            required
                            placeholder="*****"
                            class="w-full"
                        />
                    </UFormField>

                    <div>
                        <UButton
                            type="submit"
                            block
                            :loading="isLoading"
                            color="primary"
                            variant="solid"
                            class="cursor-pointer"
                        >
                            Sign in
                        </UButton>
                    </div>
                </form>
            </div>
        </UCard>
    </div>
</template>

<script setup>
const username = ref("");
const password = ref("");
const error = ref("");
const isLoading = ref(false);

const toast = useToast();

const config = useRuntimeConfig();

const handleSubmit = async (event) => {
    event.preventDefault();

    try {
        error.value = null;
        isLoading.value = true;
        fetch(`${config.public.apiBase}/auth/login`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                username: username.value,
                password: password.value,
            }),
        }).then((response) => {
            if (response.status != 200) {
                toast.add({
                    title: "Login Failed",
                    description: response.body,
                    color: "error",
                });
            } else {
                toast.add({
                    title: "Login Successful",
                    description: "You have been successfully logged in.",
                    color: "success",
                });
            }
        });
    } catch (err) {
        console.error("Login failed:", err);
        error.value =
            err.response?.data?.message || "An error occurred during login";

        toast.add({
            title: "Login Failed",
            description: error.value,
            color: "error",
        });
    } finally {
        isLoading.value = false;
    }
};
</script>
