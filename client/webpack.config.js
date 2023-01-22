const createExpoWebpackConfigAsync = require("@expo/webpack-config");

module.exports = async function (env, argv) {
    const config = await createExpoWebpackConfigAsync(
        {
            ...env,
            test: /\.(js|jsx)$/,
            babel: {
                dangerouslyAddModulePathsToTranspile: ["@reach/combobox"],
            },
        },
        argv
    );
    // Customize the config before returning it.
    return config;
};
